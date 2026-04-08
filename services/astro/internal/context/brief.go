package context

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Camionerou/rag-saldivia/services/astro/internal/astromath"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/ephemeris"
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

	// Section 5: Transits
	b.WriteString("## TRÁNSITOS LENTOS\n\n")
	if len(ctx.Transits) > 0 {
		for _, tr := range ctx.Transits {
			retro := ""
			if tr.Retrograde {
				retro = " Rx"
			}
			b.WriteString(fmt.Sprintf("- %s %s %s (orbe %.2f° %d pasadas%s mes %d %s)\n",
				tr.Transit, tr.Aspect, tr.Natal, tr.Orb, tr.Passes, retro, tr.Month, tr.Nature))
		}
	} else {
		b.WriteString("Sin tránsitos lentos activos.\n")
	}
	b.WriteString("\n")

	// Section 5b: Stations
	if len(ctx.Stations) > 0 {
		b.WriteString("## ESTACIONES PLANETARIAS\n\n")
		for _, st := range ctx.Stations {
			line := fmt.Sprintf("- %s %s en %s (mes %d, día %d)",
				st.Planet, st.Type, astromath.PosToStr(st.Lon), st.Month, st.Day)
			if st.NatPoint != "" {
				line += fmt.Sprintf(" ⚠ cerca de %s (orbe %.1f°)", st.NatPoint, st.NatOrb)
			}
			b.WriteString(line + "\n")
		}
		b.WriteString("\n")
	}

	// Section 6: Eclipses
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

	// Section 7: Dignities and Disposition (Plan 12)
	if ctx.Almuten != nil {
		b.WriteString("## DIGNIDADES Y DISPOSICIÓN\n\n")
		b.WriteString(fmt.Sprintf("**Almutén Figuris:** %s (puntaje %d)\n", ctx.Almuten.Winner, ctx.Almuten.Score))
		if ctx.Disposition != nil {
			if ctx.Disposition.FinalDispositor != "" {
				b.WriteString(fmt.Sprintf("**Dispositor final:** %s\n", ctx.Disposition.FinalDispositor))
			}
			if len(ctx.Disposition.MutualReceptions) > 0 {
				for _, mr := range ctx.Disposition.MutualReceptions {
					b.WriteString(fmt.Sprintf("  Recepción mutua: %s (%s) ↔ %s (%s)\n",
						mr.PlanetA, mr.SignA, mr.PlanetB, mr.SignB))
				}
			}
		}
		if ctx.Sect != nil {
			b.WriteString(fmt.Sprintf("**Secta:** luz=%s, benéfico=%s, maléfico=%s\n",
				ctx.Sect.SectLight, ctx.Sect.SectBenefic, ctx.Sect.SectMalefic))
			if len(ctx.Sect.Hayz) > 0 {
				b.WriteString(fmt.Sprintf("  Hayz: %s\n", strings.Join(ctx.Sect.Hayz, ", ")))
			}
		}
		b.WriteString("\n")
	}

	// Section 8: Decennials (Plan 12)
	if ctx.Decennials != nil {
		b.WriteString("## DECENIALES\n\n")
		b.WriteString(fmt.Sprintf("**Período mayor:** %s (%d años, hasta edad %.1f)\n",
			ctx.Decennials.MajorPlanet, ctx.Decennials.MajorYears, ctx.Decennials.MajorEndAge))
		b.WriteString(fmt.Sprintf("**Sub-período:** %s (edad %.1f–%.1f)\n",
			ctx.Decennials.SubPlanet, ctx.Decennials.SubStartAge, ctx.Decennials.SubEndAge))
		b.WriteString(fmt.Sprintf("**Próximo mayor:** %s\n\n", ctx.Decennials.NextMajor))
	}

	// Section 9: Tertiary Progressions (Plan 12)
	if ctx.TertiaryProg != nil && len(ctx.TertiaryProg.Positions) > 0 {
		b.WriteString("## PROGRESIONES TERCIARIAS\n\n")
		for _, pp := range ctx.TertiaryProg.Positions {
			line := fmt.Sprintf("- %s TP en %s (casa %d)", pp.Name, pp.Sign, pp.House)
			if pp.Retro {
				line += " Rx"
			}
			if pp.SignIngress {
				line += fmt.Sprintf(" ⚠ INGRESO %s → %s", pp.PrevSign, pp.Sign)
			}
			b.WriteString(line + "\n")
		}
		b.WriteString("\n")
	}

	// Section 10: Fast Transits (Plan 12)
	if len(ctx.FastTransits) > 0 {
		b.WriteString("## TRÁNSITOS RÁPIDOS\n\n")
		for _, ft := range ctx.FastTransits {
			b.WriteString(fmt.Sprintf("- %s %s %s (orbe %.2f° mes %d %s)\n",
				ft.Transit, ft.Aspect, ft.Natal, ft.Orb, ft.Month, ft.Nature))
		}
		b.WriteString("\n")
	}

	// Section 11: Lunations (Plan 12)
	if ctx.Lunations != nil && len(ctx.Lunations.Lunations) > 0 {
		b.WriteString("## LUNACIONES\n\n")
		for _, l := range ctx.Lunations.Lunations {
			line := fmt.Sprintf("- Luna %s: %d/%d en %s (casa %d)",
				l.Type, l.Day, l.Month, l.Sign, l.House)
			if l.NatPoint != "" {
				line += fmt.Sprintf(" → %s %s (orbe %.1f°)", l.NatAspect, l.NatPoint, l.NatOrb)
			}
			b.WriteString(line + "\n")
		}
		b.WriteString("\n")
	}

	// Section 12: Planetary Cycles (Plan 12)
	if len(ctx.PlanetaryCycles) > 0 {
		b.WriteString("## CICLOS PLANETARIOS\n\n")
		for _, c := range ctx.PlanetaryCycles {
			b.WriteString(fmt.Sprintf("- %s (orbe %.1f° mes %d)\n", c.Description, c.Orb, c.Month))
		}
		b.WriteString("\n")
	}

	// Section 13: Eclipse Triggers (Plan 12)
	if len(ctx.EclipseTriggers) > 0 {
		b.WriteString("## ACTIVACIÓN DE ECLIPSES\n\n")
		for _, t := range ctx.EclipseTriggers {
			line := fmt.Sprintf("- %s %s eclipse %s (orbe %.1f° mes %d)",
				t.TriggerPlanet, t.Aspect, t.EclipseType, t.Orb, t.Month)
			if t.NatPoint != "" {
				line += fmt.Sprintf(" ⚠ sobre %s natal", t.NatPoint)
			}
			b.WriteString(line + "\n")
		}
		b.WriteString("\n")
	}

	// Section 14: Lots (Plan 12)
	if len(ctx.Lots) > 0 {
		b.WriteString("## LOTES HELENÍSTICOS\n\n")
		for _, lot := range ctx.Lots {
			b.WriteString(fmt.Sprintf("- %s: %s (casa %d) — %s\n",
				lot.Name, lot.Pos, lot.House, lot.Description))
		}
		if len(ctx.LotsActivated) > 0 {
			b.WriteString("\nActivaciones SA sobre lotes:\n")
			for _, la := range ctx.LotsActivated {
				b.WriteString(fmt.Sprintf("  SA %s %s %s (orbe %.1f°)\n",
					la.SAPlanet, la.Aspect, la.LotName, la.Orb))
			}
		}
		b.WriteString("\n")
	}

	// Section 15: Midpoints (Plan 12)
	if ctx.Midpoints != nil && len(ctx.Midpoints.Activated) > 0 {
		b.WriteString("## PUNTOS MEDIOS ACTIVADOS (Ebertin)\n\n")
		top := ctx.Midpoints.Activated
		if len(top) > 10 {
			top = top[:10]
		}
		for _, mp := range top {
			b.WriteString(fmt.Sprintf("- %s (orbe %.2f°)\n", mp.Notation, mp.Orb))
		}
		b.WriteString("\n")
	}

	// Section 16: Declinations (Plan 12)
	if ctx.Declinations != nil {
		if len(ctx.Declinations.OOBPlanets) > 0 {
			b.WriteString("## DECLINACIONES\n\n")
			b.WriteString("Planetas fuera de límites:\n")
			for _, p := range ctx.Declinations.OOBPlanets {
				hemisphere := "N"
				if p.Dec < 0 {
					hemisphere = "S"
				}
				b.WriteString(fmt.Sprintf("  %s: %.2f° %s ⚠ OOB\n", p.Planet, p.Dec, hemisphere))
			}
			b.WriteString("\n")
		}
		if len(ctx.Declinations.Parallels) > 0 {
			if len(ctx.Declinations.OOBPlanets) == 0 {
				b.WriteString("## DECLINACIONES\n\n")
			}
			for _, pa := range ctx.Declinations.Parallels {
				b.WriteString(fmt.Sprintf("  %s %s %s (orbe %.2f°)\n",
					pa.PlanetA, pa.Type, pa.PlanetB, pa.Orb))
			}
			b.WriteString("\n")
		}
	}

	// Section 17: Activation Chains (Plan 12)
	if len(ctx.ActivationChains) > 0 {
		b.WriteString("## CADENAS DE ACTIVACIÓN\n\n")
		for _, ac := range ctx.ActivationChains {
			b.WriteString(fmt.Sprintf("- %s — %s (%d técnicas, %s)\n",
				ac.Planet, ac.Description, ac.Count, ac.Significance))
		}
		b.WriteString("\n")
	}

	// Section 18: Transit context builders (antiscia, fixed stars, cazimi)
	b.WriteString(AntisciaContext(ctx.Chart, ctx.Year))
	b.WriteString(FixedStarsTransitContext(ctx.Year))
	b.WriteString(CazimiCombustionTransitContext(ctx.Year))

	// Section 19: Natal sub-analyses
	natalText := astromath.FormatNatalAnalysis(
		ctx.AspectPatterns, ctx.ChartShape, ctx.Hemispheres,
		ctx.FullDignities, ctx.PlanetaryAge,
	)
	if natalText != "" {
		b.WriteString(natalText)
	}

	// Section 19: Cross-technique analyses
	crossText := FormatCrossAnalyses(
		ctx.RSLRCrossings, ctx.PrenatalTransits,
		ctx.Divisor, ctx.TriplicityLords, ctx.ChronoCross,
	)
	if crossText != "" {
		b.WriteString(crossText)
	}

	// Section 20: Scoring + Synthesis
	b.WriteString(fmt.Sprintf("## SCORE DE ACTIVACIÓN: %d/100\n\n", ctx.Score))

	// Dominant themes
	themes := DominantThemes(ctx, 3)
	if len(themes) > 0 {
		b.WriteString("## TEMAS DOMINANTES\n\n")
		for _, t := range themes {
			b.WriteString(fmt.Sprintf("- %s\n", t))
		}
		b.WriteString("\n")
	}

	// Synthesis (verdicts + contradictions + month scores)
	b.WriteString(SynthesisBrief(ctx.Verdicts, ctx.Contradictions, ctx.MonthlyScores))
	b.WriteString("\n")

	// Section 21: Convergence matrix
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
// Weights: PD=3, eclipse=2, progression ingress=2 (month of birthday only),
// SA=1 (month of birthday only as background), lunar return=0 (noise).
func buildConvergenceMatrix(ctx *FullContext) []MonthScore {
	scores := make([]MonthScore, 12)
	for i := range scores {
		scores[i].Month = i + 1
	}

	// Birth month for calendar-month derivation from age
	birthMonth := 1
	if ctx.Chart != nil {
		_, bm, _, _ := ephemeris.RevJul(ctx.Chart.JD)
		birthMonth = bm
	}

	// Eclipse activations score by month
	for _, e := range ctx.Eclipses {
		m := e.Eclipse.Month
		if m >= 1 && m <= 12 {
			scores[m-1].Score += 2
			scores[m-1].Techniques = append(scores[m-1].Techniques, fmt.Sprintf("eclipse_%s", e.NatPoint))
		}
	}

	// Primary Directions: convert age to calendar month using birth month
	for _, d := range ctx.Directions {
		// Age fraction → months after birthday → calendar month
		ageYears := d.AgeExact
		monthsAfterBday := (ageYears - float64(int(ageYears))) * 12
		calMonth := ((birthMonth - 1) + int(monthsAfterBday)) % 12
		if calMonth >= 0 && calMonth < 12 {
			scores[calMonth].Score += 3
			scores[calMonth].Techniques = append(scores[calMonth].Techniques,
				fmt.Sprintf("PD_%s_%s_%s", d.Promissor, d.Aspect, d.Significator))
		}
	}

	// Solar Arc: background indicator, annotate birthday month only
	if len(ctx.SolarArc) > 0 && birthMonth >= 1 && birthMonth <= 12 {
		scores[birthMonth-1].Score++
		scores[birthMonth-1].Techniques = append(scores[birthMonth-1].Techniques,
			fmt.Sprintf("%d_SA_activas", len(ctx.SolarArc)))
	}

	// Progression ingresses: score birthday month only (year-long background effect)
	if ctx.Progressions != nil && birthMonth >= 1 && birthMonth <= 12 {
		for _, pp := range ctx.Progressions.Positions {
			if pp.SignIngress {
				scores[birthMonth-1].Score += 2
				scores[birthMonth-1].Techniques = append(scores[birthMonth-1].Techniques,
					fmt.Sprintf("prog_ingress_%s", pp.Name))
			}
		}
	}

	// Stations near natal points: highest weight (most significant transit event)
	for _, st := range ctx.Stations {
		if st.NatPoint != "" && st.Month >= 1 && st.Month <= 12 {
			scores[st.Month-1].Score += 4
			scores[st.Month-1].Techniques = append(scores[st.Month-1].Techniques,
				fmt.Sprintf("STATION_%s_%s_%s", st.Planet, st.Type, st.NatPoint))
		}
	}

	// Transits: score by episode months
	for _, tr := range ctx.Transits {
		for _, ep := range tr.EpDetails {
			for m := ep.MonthStart; m <= ep.MonthEnd; m++ {
				if m >= 1 && m <= 12 {
					scores[m-1].Score += 2
					if m == ep.MonthStart {
						scores[m-1].Techniques = append(scores[m-1].Techniques,
							fmt.Sprintf("TR_%s_%s_%s", tr.Transit, tr.Aspect, tr.Natal))
					}
				}
			}
		}
	}

	// Lunar Returns: omitted from scoring (13/year = noise, not signal)

	sort.Slice(scores, func(i, j int) bool {
		return scores[i].Month < scores[j].Month
	})

	return scores
}
