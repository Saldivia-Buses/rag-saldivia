package intelligence

import "strings"

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
