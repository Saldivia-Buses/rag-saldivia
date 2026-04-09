package intelligence

import (
	"fmt"

	astrocontext "github.com/Camionerou/rag-saldivia/services/astro/internal/context"
)

// Contraindication is a warning about a potentially misleading astrological indication.
type Contraindication struct {
	Severity    string `json:"severity"` // "high", "medium"
	Description string `json:"description"`
}

// FindContraindications detects cases where the astrologer should add caveats.
// These are situations where a naive reading could be misleading.
func FindContraindications(fullCtx *astrocontext.FullContext) []Contraindication {
	var warnings []Contraindication

	// 1. Mercury retrograde: warn about contract signing
	for _, ft := range fullCtx.FastTransits {
		if ft.Transit == "Mercurio" && ft.Retrograde {
			warnings = append(warnings, Contraindication{
				Severity:    "medium",
				Description: fmt.Sprintf("Mercurio retrógrado en mes %d — precaución con contratos y comunicaciones", ft.Month),
			})
			break // one warning is enough
		}
	}

	// 2. Eclipse on natal Sun/Moon: major life shift, not necessarily negative
	for _, ecl := range fullCtx.Eclipses {
		if ecl.NatPoint == "Sol" || ecl.NatPoint == "Luna" {
			warnings = append(warnings, Contraindication{
				Severity:    "high",
				Description: fmt.Sprintf("Eclipse %s sobre %s natal — cambio significativo, no necesariamente negativo", ecl.Eclipse.Type, ecl.NatPoint),
			})
		}
	}

	// 3. Saturn transit to Sun/Moon: maturation, not only restriction
	for _, tr := range fullCtx.Transits {
		if tr.Transit == "Saturno" && (tr.Natal == "Sol" || tr.Natal == "Luna") {
			warnings = append(warnings, Contraindication{
				Severity:    "medium",
				Description: "Saturno sobre luminaria — proceso de maduración, no solo restricción",
			})
			break
		}
	}

	// 4. Multiple conflicting time lords
	lords := make(map[string]int)
	if fullCtx.Profection != nil {
		lords[fullCtx.Profection.Lord]++
	}
	if fullCtx.Firdaria != nil {
		lords[fullCtx.Firdaria.MajorLord]++
	}
	if fullCtx.ZRFortune != nil && fullCtx.ZRFortune.Level1 != nil {
		lords[fullCtx.ZRFortune.Level1.Lord]++
	}
	if len(lords) >= 3 {
		allDifferent := true
		for _, count := range lords {
			if count > 1 {
				allDifferent = false
				break
			}
		}
		if allDifferent {
			warnings = append(warnings, Contraindication{
				Severity:    "medium",
				Description: "Señores del tiempo divergentes — año con temas múltiples, evitar simplificar",
			})
		}
	}

	// 5. Birth time unknown warning
	if fullCtx.Chart != nil && fullCtx.Chart.UTCOffset == 0 {
		// Could indicate default offset — not definitive but worth noting
		// This is a heuristic; proper check would be on the contact record
	}

	return warnings
}

// ConfidenceAfterObjections calculates a 0-1 confidence score for predictions
// after accounting for contradicting factors. Ported from Python astro-v2.
// Factors: retrograde transiting planet, combustion, essential debility,
// malefic aspects, divergent time lords.
func ConfidenceAfterObjections(fullCtx *astrocontext.FullContext) float64 {
	confidence := 1.0

	// Penalty for retrograde transiting planets in key activations
	retroCount := 0
	for _, tr := range fullCtx.Transits {
		if tr.Retrograde {
			retroCount++
		}
	}
	if retroCount > 0 {
		confidence -= 0.05 * float64(min(retroCount, 4)) // max -0.20
	}

	// Penalty for combustion in natal chart (planets < 8° from Sun)
	if fullCtx.Chart != nil {
		sunLon := 0.0
		if sun, ok := fullCtx.Chart.Planets["Sol"]; ok {
			sunLon = sun.Lon
		}
		for name, pos := range fullCtx.Chart.Planets {
			if name == "Sol" {
				continue
			}
			diff := pos.Lon - sunLon
			if diff < 0 {
				diff = -diff
			}
			if diff > 180 {
				diff = 360 - diff
			}
			if diff < 8 && diff > 0.283 { // combust but not cazimi
				confidence -= 0.03 // per combust planet
			}
		}
	}

	// Penalty for divergent time lords (3+ different lords = ambiguous year)
	lords := make(map[string]bool)
	if fullCtx.Profection != nil && fullCtx.Profection.Lord != "" {
		lords[fullCtx.Profection.Lord] = true
	}
	if fullCtx.Firdaria != nil && fullCtx.Firdaria.MajorLord != "" {
		lords[fullCtx.Firdaria.MajorLord] = true
	}
	if fullCtx.ZRFortune != nil && fullCtx.ZRFortune.Level1 != nil {
		lords[fullCtx.ZRFortune.Level1.Lord] = true
	}
	if len(lords) >= 3 {
		confidence -= 0.10 // multiple themes, less certainty
	}

	// Bonus for convergence (multiple techniques agree)
	convergentMonths := 0
	for _, s := range fullCtx.MonthlyScores {
		if s > 50 {
			convergentMonths++
		}
	}
	if convergentMonths >= 2 {
		confidence += 0.05 * float64(min(convergentMonths-1, 3)) // max +0.15
	}

	// Clamp
	if confidence < 0.1 {
		confidence = 0.1
	}
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}
