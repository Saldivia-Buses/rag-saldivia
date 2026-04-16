package spine

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	natspub "github.com/Camionerou/rag-saldivia/pkg/nats"
)

// placeholderRe matches {name} segments in a subject template.
var placeholderRe = regexp.MustCompile(`\{([a-zA-Z_][a-zA-Z0-9_]*)\}`)

// BuildSubject substitutes {key} placeholders in template with values from
// replacements, validates every resulting segment as a NATS subject token,
// and returns the canonical subject.
//
// Each substitution value must satisfy natspub.IsValidSubjectToken
// (alphanumeric + underscore + hyphen). Literal segments in the template may
// contain dots (they split into multiple segments) but no other specials.
func BuildSubject(template string, replacements map[string]string) (string, error) {
	if template == "" {
		return "", errors.New("spine: subject template is empty")
	}

	var (
		missing []string
		badVals []string
	)
	substituted := placeholderRe.ReplaceAllStringFunc(template, func(match string) string {
		key := match[1 : len(match)-1]
		val, ok := replacements[key]
		if !ok {
			missing = append(missing, key)
			return match
		}
		if !natspub.IsValidSubjectToken(val) {
			badVals = append(badVals, fmt.Sprintf("%s=%q", key, val))
			return match
		}
		return val
	})
	if len(missing) > 0 {
		return "", fmt.Errorf("spine: unknown placeholder(s) in subject template: %s", strings.Join(missing, ", "))
	}
	if len(badVals) > 0 {
		return "", fmt.Errorf("spine: invalid placeholder value(s): %s", strings.Join(badVals, ", "))
	}

	if err := ValidateSubject(substituted); err != nil {
		return "", err
	}
	return substituted, nil
}

// ValidateSubject returns nil iff s is a canonical dot-separated NATS subject
// where every segment is a valid natspub subject token (alphanumeric,
// underscore, hyphen). Empty segments (leading/trailing/double dots) are
// rejected.
func ValidateSubject(s string) error {
	if s == "" {
		return errors.New("spine: subject is empty")
	}
	for i, segment := range strings.Split(s, ".") {
		if !natspub.IsValidSubjectToken(segment) {
			return fmt.Errorf("spine: invalid subject segment at position %d: %q", i, segment)
		}
	}
	return nil
}
