package intelligence

// AdaptiveConfig holds runtime parameters adjusted by query complexity.
type AdaptiveConfig struct {
	ThinkingBudget int    `json:"thinking_budget"` // tokens for extended thinking
	MaxChars       int    `json:"max_chars"`       // max response length
	Model          string `json:"model"`           // "opus", "sonnet"
	Depth          int    `json:"depth"`           // 1-5 (flash to exhaustive)
}

// AdaptiveThinking adjusts LLM parameters based on query complexity.
// Complexity factors: number of techniques, cross-references, domain depth.
func AdaptiveThinking(
	techniqueCount int,
	crossRefCount int,
	domain *ResolvedDomain,
	briefLength int,
) *AdaptiveConfig {
	cfg := &AdaptiveConfig{
		ThinkingBudget: 8000,
		MaxChars:       5000,
		Model:          "opus",
		Depth:          3,
	}

	// Simple query (natal, few techniques)
	if techniqueCount < 5 && crossRefCount == 0 {
		cfg.ThinkingBudget = 4000
		cfg.MaxChars = 3000
		cfg.Depth = 2
	}

	// Complex query (many techniques, cross-refs, enterprise)
	if techniqueCount > 15 || crossRefCount > 3 {
		cfg.ThinkingBudget = 16000
		cfg.MaxChars = 8000
		cfg.Depth = 4
	}

	// Enterprise queries need more depth
	if domain != nil && (domain.ID == "empresa" || domain.Parent == "empresa") {
		cfg.ThinkingBudget = 16000
		cfg.MaxChars = 10000
		cfg.Depth = 5
	}

	// Brief too large — increase thinking budget to handle it
	if briefLength > 10000 {
		cfg.ThinkingBudget = 16000
	}

	return cfg
}

// PrescoredConvergence does a lightweight pre-scan to identify months
// with high convergence before running full calculations.
// Useful for optimizing which months to analyze in detail.
type PrescoredMonth struct {
	Month    int     `json:"month"`
	Score    float64 `json:"score"` // 0-1 preliminary score
	Sources  int     `json:"sources"` // distinct technique types contributing
}

func PrescoreConvergence(
	transits int, // transit episode count
	eclipses int, // eclipse activation count
	stations int, // station count near natal points
) [12]PrescoredMonth {
	var months [12]PrescoredMonth
	for i := range months {
		months[i].Month = i + 1
	}

	// Distribute scores proportionally (simplified pre-scan)
	// In production, this would parse the raw context text like Python does
	totalEvents := transits + eclipses + stations
	if totalEvents == 0 {
		return months
	}

	// Even distribution as baseline — real implementation would parse actual months
	avgPerMonth := float64(totalEvents) / 12.0
	for i := range months {
		months[i].Score = avgPerMonth / float64(totalEvents)
		months[i].Sources = 1
		if avgPerMonth > 3 { months[i].Sources = 2 }
		if avgPerMonth > 6 { months[i].Sources = 3 }
	}

	return months
}
