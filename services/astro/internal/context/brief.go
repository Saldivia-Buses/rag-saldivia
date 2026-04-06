package context

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Camionerou/rag-saldivia/services/astro/internal/astromath"
)

// MonthScore holds convergence data for a single month.
type MonthScore struct {
	Month      int      `json:"month"`
	Score      int      `json:"score"`
	Techniques []string `json:"techniques"`
}

// BuildBrief produces a structured intelligence brief from the full context.
// This text becomes the LLM system prompt for narration.
func BuildBrief(ctx *FullContext) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("# BRIEF DE INTELIGENCIA ASTROLÓGICA — %s — %d\n\n", ctx.ContactName, ctx.Year))

	// Section 1: Time Lords (who rules the year)
	b.WriteString("## SEÑORES DEL TIEMPO\n\n")
	if ctx.Profection != nil {
		b.WriteString(fmt.Sprintf("**Profección anual:** Casa %d activa, signo %s, cronócrata: %s\n",
			ctx.Profection.ActiveHouse, ctx.Profection.ProfSign, ctx.Profection.Lord))
		b.WriteString(fmt.Sprintf("  Tema de casa: %s\n\n", ctx.Profection.Theme))
	}
	if ctx.Firdaria != nil {
		b.WriteString(fmt.Sprintf("**Firdaria:** período mayor %s (%d años), sub-período %s\n\n",
			ctx.Firdaria.MajorLord, ctx.Firdaria.MajorYears, ctx.Firdaria.SubLord))
	}
	if ctx.ZRFortune != nil && ctx.ZRFortune.Level1 != nil {
		b.WriteString(fmt.Sprintf("**ZR Fortuna:** L1=%s(%s)", ctx.ZRFortune.Level1.Sign, ctx.ZRFortune.Level1.Lord))
		if ctx.ZRFortune.Level2 != nil {
			b.WriteString(fmt.Sprintf(", L2=%s(%s)", ctx.ZRFortune.Level2.Sign, ctx.ZRFortune.Level2.Lord))
		}
		if ctx.ZRFortune.Level1.Loosing {
			b.WriteString(" ⚠ LOOSING OF THE BOND")
		}
		b.WriteString("\n")
	}
	if ctx.ZRSpirit != nil && ctx.ZRSpirit.Level1 != nil {
		b.WriteString(fmt.Sprintf("**ZR Espíritu:** L1=%s(%s)\n", ctx.ZRSpirit.Level1.Sign, ctx.ZRSpirit.Level1.Lord))
	}
	b.WriteString("\n")

	// Section 2: Primary Directions (highest precision)
	b.WriteString("## DIRECCIONES PRIMARIAS (precisión: meses)\n\n")
	if len(ctx.Directions) > 0 {
		top := min(10, len(ctx.Directions))
		for _, d := range ctx.Directions[:top] {
			applying := "separando"
			if d.Applying {
				applying = "aplicando"
			}
			b.WriteString(fmt.Sprintf("- %s %s %s — arco %.2f° edad %.1f (orbe %.2f° %s %s)\n",
				d.Promissor, d.Aspect, d.Significator, d.Arc, d.AgeExact, d.OrbDeg, d.Tipo, applying))
		}
	} else {
		b.WriteString("Sin activaciones dentro del orbe.\n")
	}
	b.WriteString("\n")

	// Section 3: Solar Arc
	b.WriteString("## ARCOS SOLARES\n\n")
	if len(ctx.SolarArc) > 0 {
		for _, sa := range ctx.SolarArc {
			b.WriteString(fmt.Sprintf("- SA %s %s %s (orbe %.2f° %s)\n",
				sa.SAplanet, sa.Aspect, sa.NatPlanet, sa.Orb, sa.Nature))
		}
	} else {
		b.WriteString("Sin activaciones dentro del orbe.\n")
	}
	b.WriteString("\n")

	// Section 4: Progressions
	b.WriteString("## PROGRESIONES SECUNDARIAS\n\n")
	if ctx.Progressions != nil {
		for _, pp := range ctx.Progressions.Positions {
			line := fmt.Sprintf("- %s prog en %s (casa %d)", pp.Name, pp.Sign, pp.House)
			if pp.Retro {
				line += " Rx"
			}
			if pp.SignIngress {
				line += fmt.Sprintf(" ⚠ INGRESO de signo (%s → %s)", pp.PrevSign, pp.Sign)
			}
			if pp.HouseIngress {
				line += fmt.Sprintf(" ⚠ INGRESO de casa (%d → %d)", pp.PrevHouse, pp.House)
			}
			b.WriteString(line + "\n")
		}
	}
	b.WriteString("\n")

	// Section 5: Eclipses
	b.WriteString("## ECLIPSES\n\n")
	if len(ctx.Eclipses) > 0 {
		for _, e := range ctx.Eclipses {
			b.WriteString(fmt.Sprintf("- Eclipse %s %s → %s %s (orbe %.2f°, mes %d)\n",
				e.Eclipse.Type, e.Eclipse.SubType, e.Aspect, e.NatPoint, e.Orb, e.Eclipse.Month))
		}
	} else {
		b.WriteString("Sin activaciones de eclipses sobre puntos natales.\n")
	}
	b.WriteString("\n")

	// Section 6: Solar Return summary
	b.WriteString("## REVOLUCIÓN SOLAR\n\n")
	if ctx.SolarReturn != nil {
		b.WriteString(fmt.Sprintf("ASC RS: %s, MC RS: %s\n",
			astromath.PosToStr(ctx.SolarReturn.ASC), astromath.PosToStr(ctx.SolarReturn.MC)))
	}
	b.WriteString("\n")

	// Section 7: Convergence matrix
	b.WriteString("## MATRIZ DE CONVERGENCIA MENSUAL\n\n")
	scores := buildConvergenceMatrix(ctx)
	for _, ms := range scores {
		bar := strings.Repeat("█", ms.Score)
		b.WriteString(fmt.Sprintf("  %2d | %s %d — %s\n", ms.Month, bar, ms.Score, strings.Join(ms.Techniques, ", ")))
	}
	b.WriteString("\n")

	return b.String()
}

// buildConvergenceMatrix scores each month based on technique overlap.
func buildConvergenceMatrix(ctx *FullContext) []MonthScore {
	scores := make([]MonthScore, 12)
	for i := range scores {
		scores[i].Month = i + 1
	}

	// Eclipse activations score by month
	for _, e := range ctx.Eclipses {
		m := e.Eclipse.Month
		if m >= 1 && m <= 12 {
			scores[m-1].Score += 2
			scores[m-1].Techniques = append(scores[m-1].Techniques, fmt.Sprintf("eclipse_%s", e.NatPoint))
		}
	}

	// Primary Directions: approximate month from age
	for _, d := range ctx.Directions {
		fractionalYear := d.AgeExact - float64(int(d.AgeExact))
		month := int(fractionalYear*12) + 1
		if month >= 1 && month <= 12 {
			scores[month-1].Score += 3 // PD is highest weight
			scores[month-1].Techniques = append(scores[month-1].Techniques,
				fmt.Sprintf("PD_%s_%s_%s", d.Promissor, d.Aspect, d.Significator))
		}
	}

	// Solar Arc activations (year-long, score all months equally but +1)
	if len(ctx.SolarArc) > 0 {
		for i := range scores {
			scores[i].Score++
			if i == 0 { // only annotate once
				scores[i].Techniques = append(scores[i].Techniques, fmt.Sprintf("%d_SA_activas", len(ctx.SolarArc)))
			}
		}
	}

	// Lunar Returns: score their months
	for _, lr := range ctx.LunarReturns {
		if lr.Month >= 1 && lr.Month <= 12 {
			scores[lr.Month-1].Score++
			scores[lr.Month-1].Techniques = append(scores[lr.Month-1].Techniques, "lunar_return")
		}
	}

	// Progressions ingresses: score all months (year-long effect)
	if ctx.Progressions != nil {
		for _, pp := range ctx.Progressions.Positions {
			if pp.SignIngress {
				for i := range scores {
					scores[i].Score += 2
					if i == 0 {
						scores[i].Techniques = append(scores[i].Techniques,
							fmt.Sprintf("prog_ingress_%s", pp.Name))
					}
				}
			}
		}
	}

	sort.Slice(scores, func(i, j int) bool {
		return scores[i].Month < scores[j].Month
	})

	return scores
}
