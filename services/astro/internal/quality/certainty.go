package quality

import (
	"github.com/Camionerou/rag-saldivia/services/astro/internal/intelligence"
)

// CertaintyResult holds the confidence scoring for a prediction.
type CertaintyResult struct {
	Score       float64 `json:"score"`       // 0.0-1.0
	Level       string  `json:"level"`       // "alta", "media", "baja"
	Factors     int     `json:"factors"`     // number of supporting techniques
	CrossRefs   int     `json:"cross_refs"`  // number of cross-references
	Description string  `json:"description"` // human-readable explanation
}

// ScoreCertainty rates prediction certainty based on technique convergence.
// More cross-references + more validated techniques = higher certainty.
func ScoreCertainty(crossRefs []intelligence.CrossReference, gate *intelligence.GateResult) *CertaintyResult {
	result := &CertaintyResult{}

	// Base: number of validated techniques (normalized)
	validatedCount := len(gate.Validated)
	result.Factors = validatedCount

	// Score from technique coverage
	techScore := 0.0
	if validatedCount >= 15 {
		techScore = 0.4
	} else if validatedCount >= 10 {
		techScore = 0.3
	} else if validatedCount >= 5 {
		techScore = 0.2
	} else {
		techScore = 0.1
	}

	// Score from cross-references (most important — convergence = confidence)
	result.CrossRefs = len(crossRefs)
	crossScore := 0.0
	for _, cr := range crossRefs {
		crossScore += cr.Significance * 0.15 // each high-significance ref adds ~0.15
	}
	if crossScore > 0.5 {
		crossScore = 0.5
	}

	// Score from gate coverage
	coverageScore := gate.Coverage * 0.1

	result.Score = techScore + crossScore + coverageScore
	if result.Score > 1.0 {
		result.Score = 1.0
	}

	// Level classification
	switch {
	case result.Score >= 0.7:
		result.Level = "alta"
		result.Description = "Alta convergencia de técnicas — predicción bien sustentada"
	case result.Score >= 0.4:
		result.Level = "media"
		result.Description = "Convergencia moderada — considerar con cautela"
	default:
		result.Level = "baja"
		result.Description = "Pocos indicadores convergentes — orientativo únicamente"
	}

	return result
}
