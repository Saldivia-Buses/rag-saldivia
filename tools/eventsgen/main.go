// Command eventsgen reads event specs from CUE files in pkg/events/spec and
// emits typed Go structs, TypeScript interfaces, and Markdown documentation.
//
// Usage:
//
//	eventsgen -spec ./pkg/events/spec \
//	          -out-go ./pkg/events/gen \
//	          -out-ts ./apps/web/src/lib/events/gen \
//	          -out-docs ./docs/events
//
// Design notes:
//   - Each *.cue file is one "family" named after the basename (notify.cue → notify).
//   - All files share package events; the cuelang.org/go runtime merges them into
//     a single value, so we use the AST to determine which Types live in each file.
//   - The payload shape is extracted from the CUE value tree via Kind + field
//     iteration. Disjunctions of string literals become enum constants in Go/TS.
package main

import (
	"bytes"
	"embed"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/ast"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	"cuelang.org/go/cue/parser"
)

//go:embed templates/*.tmpl
var templatesFS embed.FS

// ----- model -----

type Family struct {
	Name   string
	Events []Event
}

type Event struct {
	Type       string // "chat.new_message"
	GoName     string // "ChatNewMessage"
	Version    int
	Subject    string // template with {slug}, {action}, ...
	Payload    []Field
	Publishers []string
	Consumers  []string
}

type Field struct {
	Name       string // JSON name: "user_id"
	GoName     string // "UserID"
	TSName     string // "userId"
	GoType     string
	TSType     string
	Optional   bool
	EnumValues []string
	EnumGoType string // set when EnumValues non-empty, e.g. "ChatNewMessageChannel"
}

// ----- main -----

func main() {
	specDir := flag.String("spec", "pkg/events/spec", "CUE spec directory")
	outGo := flag.String("out-go", "pkg/events/gen", "Go output directory")
	outTS := flag.String("out-ts", "apps/web/src/lib/events/gen", "TypeScript output directory")
	outDocs := flag.String("out-docs", "docs/events", "Markdown output directory")
	flag.Parse()

	families, err := loadFamilies(*specDir)
	if err != nil {
		log.Fatalf("eventsgen: %v", err)
	}

	for _, d := range []string{*outGo, *outTS, *outDocs} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			log.Fatalf("mkdir %s: %v", d, err)
		}
	}

	for _, fam := range families {
		goDir := filepath.Join(*outGo, fam.Name)
		if err := os.MkdirAll(goDir, 0o755); err != nil {
			log.Fatalf("mkdir go family: %v", err)
		}
		if err := render(fam, "go.tmpl", filepath.Join(goDir, fam.Name+".go")); err != nil {
			log.Fatalf("render go %s: %v", fam.Name, err)
		}
		if err := render(fam, "ts.tmpl", filepath.Join(*outTS, fam.Name+".ts")); err != nil {
			log.Fatalf("render ts %s: %v", fam.Name, err)
		}
		if err := render(fam, "docs.tmpl", filepath.Join(*outDocs, fam.Name+".md")); err != nil {
			log.Fatalf("render docs %s: %v", fam.Name, err)
		}
	}

	fmt.Fprintf(os.Stderr, "eventsgen: %d families generated\n", len(families))
}

// ----- CUE loading -----

func loadFamilies(specDir string) ([]Family, error) {
	entries, err := os.ReadDir(specDir)
	if err != nil {
		return nil, fmt.Errorf("read spec dir: %w", err)
	}
	var cueFiles []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".cue") {
			cueFiles = append(cueFiles, filepath.Join(specDir, e.Name()))
		}
	}
	sort.Strings(cueFiles)

	cfg := &load.Config{Dir: specDir}
	insts := load.Instances([]string{"."}, cfg)
	if len(insts) == 0 {
		return nil, fmt.Errorf("no CUE instances in %s", specDir)
	}
	if insts[0].Err != nil {
		return nil, fmt.Errorf("load %s: %w", specDir, insts[0].Err)
	}
	ctx := cuecontext.New()
	rootVal := ctx.BuildInstance(insts[0])
	if err := rootVal.Err(); err != nil {
		return nil, fmt.Errorf("build instance: %w", err)
	}

	eventsVal := rootVal.LookupPath(cue.ParsePath("events"))
	if !eventsVal.Exists() {
		return nil, fmt.Errorf("no `events` field in CUE package")
	}

	typeToVal := map[string]cue.Value{}
	iter, err := eventsVal.Fields()
	if err != nil {
		return nil, fmt.Errorf("iterate events: %w", err)
	}
	for iter.Next() {
		typeToVal[unquote(iter.Selector().String())] = iter.Value()
	}

	var families []Family
	for _, path := range cueFiles {
		famName := strings.TrimSuffix(filepath.Base(path), ".cue")
		types, err := typesInFile(path)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
		fam := Family{Name: famName}
		for _, t := range types {
			val, ok := typeToVal[t]
			if !ok {
				return nil, fmt.Errorf("type %q declared in %s but missing from evaluated instance", t, path)
			}
			evt, err := buildEvent(t, val)
			if err != nil {
				return nil, fmt.Errorf("event %s in %s: %w", t, path, err)
			}
			fam.Events = append(fam.Events, evt)
		}
		sort.Slice(fam.Events, func(i, j int) bool { return fam.Events[i].Type < fam.Events[j].Type })
		families = append(families, fam)
	}
	return families, nil
}

// typesInFile returns Types declared at `events: "<type>": ...` in a CUE file.
func typesInFile(path string) ([]string, error) {
	f, err := parser.ParseFile(path, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	var types []string
	for _, decl := range f.Decls {
		field, ok := decl.(*ast.Field)
		if !ok {
			continue
		}
		if ident, ok := field.Label.(*ast.Ident); !ok || ident.Name != "events" {
			continue
		}
		// CUE label chain `events: "chat.new_message": {...}` parses as
		// Field{Label: "events", Value: StructLit{Elts: [Field{Label: "chat.new_message"}]}}.
		sl, ok := field.Value.(*ast.StructLit)
		if !ok {
			continue
		}
		for _, d := range sl.Elts {
			if lit, ok := labelStringLit(d); ok {
				types = append(types, lit)
			}
		}
	}
	return types, nil
}

func labelStringLit(d ast.Decl) (string, bool) {
	f, ok := d.(*ast.Field)
	if !ok {
		return "", false
	}
	lit, ok := f.Label.(*ast.BasicLit)
	if !ok {
		return "", false
	}
	return unquote(lit.Value), true
}

func unquote(s string) string {
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}

// ----- event extraction -----

func buildEvent(eventType string, v cue.Value) (Event, error) {
	e := Event{Type: eventType, GoName: goName(eventType)}

	ver, err := v.LookupPath(cue.ParsePath("version")).Int64()
	if err != nil {
		return e, fmt.Errorf("version: %w", err)
	}
	e.Version = int(ver)

	subj, err := v.LookupPath(cue.ParsePath("subject")).String()
	if err != nil {
		return e, fmt.Errorf("subject: %w", err)
	}
	e.Subject = subj

	e.Publishers, _ = stringList(v.LookupPath(cue.ParsePath("publishers")))
	e.Consumers, _ = stringList(v.LookupPath(cue.ParsePath("consumers")))

	payloadVal := v.LookupPath(cue.ParsePath("payload"))
	if !payloadVal.Exists() {
		return e, fmt.Errorf("payload missing")
	}
	fields, err := extractFields(payloadVal, e.GoName)
	if err != nil {
		return e, fmt.Errorf("payload fields: %w", err)
	}
	e.Payload = fields
	return e, nil
}

func stringList(v cue.Value) ([]string, error) {
	if !v.Exists() {
		return nil, nil
	}
	iter, err := v.List()
	if err != nil {
		return nil, err
	}
	var out []string
	for iter.Next() {
		s, err := iter.Value().String()
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, nil
}

func extractFields(v cue.Value, parentGo string) ([]Field, error) {
	iter, err := v.Fields(cue.Optional(true))
	if err != nil {
		return nil, err
	}
	var fields []Field
	for iter.Next() {
		name := iter.Selector().Unquoted()
		f := Field{
			Name:     name,
			GoName:   goName(name),
			TSName:   tsName(name),
			Optional: iter.IsOptional(),
		}
		fv := iter.Value()

		if enum, ok := enumStrings(fv); ok {
			f.EnumValues = enum
			f.EnumGoType = parentGo + goName(name)
			f.GoType = f.EnumGoType
			f.TSType = tsEnumUnion(enum)
		} else {
			f.GoType = goKind(fv)
			f.TSType = tsKind(fv)
		}
		fields = append(fields, f)
	}
	return fields, nil
}

// enumStrings returns the disjuncts if v is a union of string literals.
func enumStrings(v cue.Value) ([]string, bool) {
	op, args := v.Expr()
	if op != cue.OrOp {
		return nil, false
	}
	var values []string
	for _, a := range args {
		if a.Kind() != cue.StringKind {
			return nil, false
		}
		s, err := a.String()
		if err != nil {
			return nil, false
		}
		values = append(values, s)
	}
	if len(values) == 0 {
		return nil, false
	}
	return values, true
}

func goKind(v cue.Value) string {
	switch v.IncompleteKind() {
	case cue.StringKind:
		return "string"
	case cue.IntKind:
		return "int"
	case cue.NumberKind, cue.FloatKind:
		return "float64"
	case cue.BoolKind:
		return "bool"
	default:
		return "any"
	}
}

func tsKind(v cue.Value) string {
	switch v.IncompleteKind() {
	case cue.StringKind:
		return "string"
	case cue.IntKind, cue.NumberKind, cue.FloatKind:
		return "number"
	case cue.BoolKind:
		return "boolean"
	default:
		return "unknown"
	}
}

func tsEnumUnion(values []string) string {
	quoted := make([]string, len(values))
	for i, v := range values {
		quoted[i] = `"` + v + `"`
	}
	return strings.Join(quoted, " | ")
}

// ----- identifier helpers -----

// goName converts "chat.new_message" → "ChatNewMessage", "user_id" → "UserID".
func goName(s string) string {
	acrons := map[string]string{"id": "ID", "url": "URL", "usd": "USD", "ip": "IP", "ms": "MS", "ua": "UA"}
	var out strings.Builder
	for _, part := range splitOn(s, '.', '_') {
		lower := strings.ToLower(part)
		if ac, ok := acrons[lower]; ok {
			out.WriteString(ac)
			continue
		}
		if len(part) == 0 {
			continue
		}
		out.WriteString(strings.ToUpper(part[:1]))
		out.WriteString(part[1:])
	}
	return out.String()
}

// tsName converts "user_id" → "userId".
func tsName(s string) string {
	parts := splitOn(s, '_')
	if len(parts) == 0 {
		return s
	}
	var out strings.Builder
	out.WriteString(strings.ToLower(parts[0]))
	for _, p := range parts[1:] {
		if p == "" {
			continue
		}
		out.WriteString(strings.ToUpper(p[:1]))
		out.WriteString(p[1:])
	}
	return out.String()
}

func splitOn(s string, seps ...rune) []string {
	sepSet := make(map[rune]struct{}, len(seps))
	for _, c := range seps {
		sepSet[c] = struct{}{}
	}
	var parts []string
	var cur strings.Builder
	for _, r := range s {
		if _, ok := sepSet[r]; ok {
			if cur.Len() > 0 {
				parts = append(parts, cur.String())
				cur.Reset()
			}
			continue
		}
		cur.WriteRune(r)
	}
	if cur.Len() > 0 {
		parts = append(parts, cur.String())
	}
	return parts
}

// ----- rendering -----

func render(fam Family, tmplName, outPath string) error {
	raw, err := templatesFS.ReadFile("templates/" + tmplName)
	if err != nil {
		return fmt.Errorf("read template: %w", err)
	}
	tmpl, err := template.New(tmplName).Funcs(template.FuncMap{
		"join":    strings.Join,
		"goName":  goName,
		"listq":   func(xs []string) string { qs := make([]string, len(xs)); for i, x := range xs { qs[i] = `"` + x + `"` }; return strings.Join(qs, ", ") },
	}).Parse(string(raw))
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, fam); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}
	return os.WriteFile(outPath, buf.Bytes(), 0o644)
}
