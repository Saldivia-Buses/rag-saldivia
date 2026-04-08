package intelligence

import (
	"regexp"
	"strings"
)

// CompressBrief reduces a brief's token count by removing low-weight sections.
// Used when the combined prompt (system + brief + query) exceeds the LLM context window.
// Removes sections in reverse weight order (least important first).
func CompressBrief(brief string, maxChars int) string {
	if len(brief) <= maxChars {
		return brief
	}

	// Split into sections by "## " headers
	sections := splitSections(brief)
	if len(sections) <= 1 {
		return brief[:maxChars]
	}

	// Remove sections from the end (lowest priority) until under budget
	// Keep first section (header + time lords) and last section (convergence matrix) always
	for len(sections) > 2 {
		total := 0
		for _, s := range sections {
			total += len(s)
		}
		if total <= maxChars {
			break
		}
		// Remove second-to-last section (preserve convergence matrix at end)
		sections = append(sections[:len(sections)-2], sections[len(sections)-1])
	}

	return strings.Join(sections, "\n")
}

// CompressContext removes empty/ghost technique sections from a brief.
// Sections with "Sin activaciones" or "Sin datos" are stripped entirely.
func CompressContext(brief string) string {
	sections := splitSections(brief)
	var kept []string
	for _, s := range sections {
		lower := strings.ToLower(s)
		if strings.Contains(lower, "sin activaciones") && len(s) < 200 {
			continue // skip empty sections
		}
		if strings.Contains(lower, "sin datos") && len(s) < 100 {
			continue
		}
		kept = append(kept, s)
	}
	return strings.Join(kept, "\n")
}

// CompressBCA applies Brief Compacto Astrológico compression.
// Reduces format, not data: compacts headers, abbreviates aspects/planets,
// deduplicates patterns, removes decorative lines. Achieves 40-60% reduction.
// Ported from Python astro-v2 brief_compressor.py.
func CompressBCA(brief string) string {
	if len(brief) < 500 {
		return brief
	}

	lines := strings.Split(brief, "\n")
	var result []string
	seen := make(map[string]bool)
	prevEmpty := false

	for _, line := range lines {
		stripped := strings.TrimSpace(line)

		// Skip empty lines (keep max 1)
		if stripped == "" {
			if prevEmpty {
				continue
			}
			result = append(result, "")
			prevEmpty = true
			continue
		}
		prevEmpty = false

		// Skip purely decorative lines
		if reDecorative.MatchString(stripped) {
			continue
		}

		// Skip instruction lines the LLM already knows
		lowerLine := strings.ToLower(stripped)
		if containsAny(lowerLine, instructionSkips) {
			continue
		}

		// Compact section headers
		compressed := compactHeader(stripped)

		// Abbreviate aspects
		for full, abbr := range aspectAbbr {
			compressed = strings.ReplaceAll(compressed, full, abbr)
		}

		// Abbreviate planets
		for full, abbr := range planetAbbr {
			compressed = strings.ReplaceAll(compressed, full, abbr)
		}

		// Remove verbose qualifiers
		compressed = strings.ReplaceAll(compressed, "período mayor ", "")
		compressed = strings.ReplaceAll(compressed, "período menor ", "sub:")
		compressed = strings.ReplaceAll(compressed, "sub-período ", "sub:")
		compressed = strings.ReplaceAll(compressed, "Tema de casa: ", "→ ")
		compressed = strings.ReplaceAll(compressed, "Profección anual:", "Prof:")
		compressed = strings.ReplaceAll(compressed, "cronócrata: ", "lord:")
		compressed = strings.ReplaceAll(compressed, " pasadas", "x")
		compressed = strings.ReplaceAll(compressed, " favorable", " +")
		compressed = strings.ReplaceAll(compressed, " desafiante", " −")
		compressed = strings.ReplaceAll(compressed, " disolvente", " ∿")
		compressed = strings.ReplaceAll(compressed, " transformador", " ⚡")
		compressed = strings.ReplaceAll(compressed, " tenso", " −")
		compressed = strings.ReplaceAll(compressed, " mixto", " ±")
		compressed = strings.ReplaceAll(compressed, "directo ", "D/")

		// Compress orb notation: "orbe 1.23°" → "orb 1.2°"
		compressed = reOrb.ReplaceAllString(compressed, "orb ${1}.${2}°")

		// Compress applying/separating
		compressed = strings.ReplaceAll(compressed, "aplicando", "apl")
		compressed = strings.ReplaceAll(compressed, "separando", "sep")

		// Deduplicate: skip lines with same factual content (>20 chars)
		factKey := reNonAlpha.ReplaceAllString(compressed, " ")
		if len(factKey) > 20 {
			factKey60 := factKey
			if len(factKey60) > 60 {
				factKey60 = factKey60[:60]
			}
			if seen[factKey60] {
				continue
			}
			seen[factKey60] = true
		}

		result = append(result, compressed)
	}

	output := strings.Join(result, "\n")

	// Collapse multiple empty lines
	output = reMultiNewline.ReplaceAllString(output, "\n\n")

	return output
}

// Pre-compiled regexps
var (
	reDecorative   = regexp.MustCompile(`^[═─━╔╗╚╝║│┌┐└┘├┤┬┴┼\-=_*#]{3,}$`)
	reOrb          = regexp.MustCompile(`orbe?\s*(\d+)\.(\d)\d*°`)
	reNonAlpha     = regexp.MustCompile(`[^\p{L}\p{N}]+`)
	reMultiNewline = regexp.MustCompile(`\n{3,}`)

	// Header compaction patterns
	headerCompact = []struct {
		pattern *regexp.Regexp
		replace string
	}{
		{regexp.MustCompile(`## DIRECCIONES PRIMARIAS[^\n]*`), "## DP"},
		{regexp.MustCompile(`## ARCOS SOLARES[^\n]*`), "## SA"},
		{regexp.MustCompile(`## PROGRESIONES SECUNDARIAS[^\n]*`), "## Prog2"},
		{regexp.MustCompile(`## PROGRESIONES TERCIARIAS[^\n]*`), "## ProgT"},
		{regexp.MustCompile(`## REVOLUCIÓN SOLAR[^\n]*`), "## RS"},
		{regexp.MustCompile(`## RETORNO LUNAR[^\n]*`), "## RL"},
		{regexp.MustCompile(`## PROFECCIÓN ANUAL[^\n]*`), "## Prof"},
		{regexp.MustCompile(`## FIRDARIA[^\n]*`), "## Fird"},
		{regexp.MustCompile(`## ZODIACAL RELEASING[^\n]*`), "## ZR"},
		{regexp.MustCompile(`## CONVERGENCIA[^\n]*`), "## CONV"},
		{regexp.MustCompile(`## ECLIPSES[^\n]*`), "## Ecl"},
		{regexp.MustCompile(`## TRÁNSITOS LENTOS[^\n]*`), "## TR lento"},
		{regexp.MustCompile(`## TRÁNSITOS RÁPIDOS[^\n]*`), "## TR"},
		{regexp.MustCompile(`## ESTACIONES PLANETARIAS[^\n]*`), "## Estaciones"},
		{regexp.MustCompile(`## ESTRELLAS FIJAS[^\n]*`), "## Stars"},
		{regexp.MustCompile(`## DIGNIDADES Y DISPOSICIÓN[^\n]*`), "## Dignidades"},
		{regexp.MustCompile(`## ANTISCIA[^\n]*`), "## Antiscia"},
		{regexp.MustCompile(`## SEÑORES DEL TIEMPO[^\n]*`), "## TimeLords"},
		{regexp.MustCompile(`## LOTES HELÉNISTICOS[^\n]*`), "## Lotes"},
		{regexp.MustCompile(`## PUNTOS MEDIOS[^\n]*`), "## Midpoints"},
		{regexp.MustCompile(`## DECLINACIONES[^\n]*`), "## Decl"},
		{regexp.MustCompile(`## ACTIVACIONES CRUZADAS[^\n]*`), "## XRef"},
		{regexp.MustCompile(`## CADENAS DE ACTIVACIÓN[^\n]*`), "## Chains"},
		{regexp.MustCompile(`## VENTANAS DE TIMING[^\n]*`), "## Timing"},
		{regexp.MustCompile(`## MATRIZ DE CONVERGENCIA[^\n]*`), "## Matrix"},
	}

	aspectAbbr = map[string]string{
		"conjunción": "☌", "Conjunción": "☌", "CONJUNCIÓN": "☌",
		"oposición": "☍", "Oposición": "☍", "OPOSICIÓN": "☍",
		"cuadratura": "□", "Cuadratura": "□", "CUADRATURA": "□",
		"trígono": "△", "Trígono": "△", "TRÍGONO": "△",
		"sextil": "⚹", "Sextil": "⚹", "SEXTIL": "⚹",
	}

	planetAbbr = map[string]string{
		"Plutón": "Plu", "Neptuno": "Nep", "Urano": "Ura", "Saturno": "Sat",
		"Júpiter": "Jup", "Marte": "Mar", "Venus": "Ven", "Mercurio": "Mer",
		"Quirón": "Qui",
	}

	instructionSkips = []string{
		"usá estos datos", "nunca ignorés", "todo lo que sigue",
		"datos calculados con swiss", "sistema topocéntrico",
		"columna vertebral", "instrucción obligatoria",
	}
)

func compactHeader(line string) string {
	for _, hc := range headerCompact {
		if hc.pattern.MatchString(line) {
			return hc.pattern.ReplaceAllString(line, hc.replace)
		}
	}
	return line
}

func containsAny(s string, substrs []string) bool {
	for _, sub := range substrs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

// splitSections splits a brief by "## " markdown headers.
func splitSections(text string) []string {
	lines := strings.Split(text, "\n")
	var sections []string
	var current strings.Builder

	for _, line := range lines {
		if strings.HasPrefix(line, "## ") && current.Len() > 0 {
			sections = append(sections, current.String())
			current.Reset()
		}
		current.WriteString(line)
		current.WriteString("\n")
	}
	if current.Len() > 0 {
		sections = append(sections, current.String())
	}

	return sections
}
