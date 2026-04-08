package context

import (
	"fmt"
	"math"
	"strings"

	"github.com/Camionerou/rag-saldivia/services/astro/internal/astromath"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/technique"
)

// ActivationScore computes a quantitative 0-100 score for a year.
// Combines all techniques with weights, dignity multipliers, sect considerations,
// convergence bonuses, and mutual reception bonuses.
func ActivationScore(ctx *FullContext) int {
	score := 0.0

	// Category weights
	weights := map[string]float64{
		"PD": 20, "SA": 18, "transit": 15, "profection": 12,
		"firdaria": 10, "eclipse": 8, "ZR": 7, "decennial": 5,
		"station": 5,
	}

	// Primary Directions — highest precision technique
	pdCount := 0
	for _, d := range ctx.Directions {
		if d.OrbDeg < 1.0 {
			pdCount++
			orbFactor := 1.0 - d.OrbDeg // tighter orb = higher score
			score += weights["PD"] * orbFactor
		}
	}
	// Cap PD contribution
	if pdScore := weights["PD"] * float64(pdCount); pdScore > 40 {
		score -= pdScore - 40
	}

	// Solar Arcs
	saCount := len(ctx.SolarArc)
	for _, sa := range ctx.SolarArc {
		orbFactor := 1.0 - (sa.Orb / 1.5) // orb / max orb
		if orbFactor < 0 { orbFactor = 0 }
		nature := 1.0
		if sa.Nature == "tenso" { nature = 0.8 } // tense activations count but slightly less
		score += weights["SA"] * orbFactor * nature * 0.5 // per activation
	}
	_ = saCount

	// Transits (slow planets)
	for _, tr := range ctx.Transits {
		planetWeight := 1.0
		switch tr.Transit {
		case "Plutón": planetWeight = 1.5
		case "Urano": planetWeight = 1.4
		case "Neptuno": planetWeight = 1.3
		case "Saturno": planetWeight = 1.2
		case "Júpiter": planetWeight = 1.0
		}
		orbFactor := 1.0 - (tr.Orb / 3.0)
		if orbFactor < 0 { orbFactor = 0 }
		score += weights["transit"] * orbFactor * planetWeight * 0.3
	}

	// Profection — active house and chronocrator
	if ctx.Profection != nil {
		score += weights["profection"]
		// Bonus if chronocrator is also active in transits
		for _, tr := range ctx.Transits {
			if tr.Natal == ctx.Profection.Lord || tr.Transit == ctx.Profection.Lord {
				score += 3 // convergence bonus
			}
		}
	}

	// Firdaria
	if ctx.Firdaria != nil {
		score += weights["firdaria"]
	}

	// Eclipses on natal points
	eclipseScore := float64(len(ctx.Eclipses)) * weights["eclipse"] * 0.5
	if eclipseScore > 16 { eclipseScore = 16 }
	score += eclipseScore

	// Zodiacal Releasing — loosing of the bond is extremely significant
	if ctx.ZRFortune != nil && ctx.ZRFortune.Level1 != nil && ctx.ZRFortune.Level1.Loosing {
		score += 15 // loosing of the bond = peak/crisis year
	}

	// Decennials
	if ctx.Decennials != nil {
		score += weights["decennial"]
	}

	// Stations near natal points
	for _, st := range ctx.Stations {
		if st.NatPoint != "" {
			score += weights["station"]
		}
	}

	// Convergence bonus: count unique natal points activated by 3+ techniques
	if len(ctx.ActivationChains) > 0 {
		for _, chain := range ctx.ActivationChains {
			if chain.Count >= 4 {
				score += 8
			} else if chain.Count >= 3 {
				score += 4
			}
		}
	}

	// Mutual reception bonus
	if ctx.Disposition != nil && len(ctx.Disposition.MutualReceptions) > 0 {
		score += float64(len(ctx.Disposition.MutualReceptions)) * 2
	}

	// Dignity multiplier: almuten planet active = amplified
	if ctx.Almuten != nil && ctx.Almuten.Winner != "" {
		for _, tr := range ctx.Transits {
			if tr.Transit == ctx.Almuten.Winner || tr.Natal == ctx.Almuten.Winner {
				score += 3
			}
		}
	}

	// Normalize to 0-100
	normalized := math.Min(score, 100)
	if normalized < 0 { normalized = 0 }
	return int(math.Round(normalized))
}

// ConvergenceScoreByPoint counts how many distinct techniques activate each natal point.
// Returns map of natal_point → count.
func ConvergenceScoreByPoint(ctx *FullContext) map[string]int {
	points := make(map[string]map[string]bool) // natal_point → set of technique names

	addPoint := func(point, technique string) {
		if point == "" { return }
		if points[point] == nil { points[point] = make(map[string]bool) }
		points[point][technique] = true
	}

	for _, sa := range ctx.SolarArc {
		addPoint(sa.NatPlanet, "SA")
	}
	for _, d := range ctx.Directions {
		if d.OrbDeg < 2.0 { addPoint(d.Significator, "PD") }
	}
	for _, tr := range ctx.Transits {
		addPoint(tr.Natal, "TR_"+tr.Transit)
	}
	for _, ecl := range ctx.Eclipses {
		addPoint(ecl.NatPoint, "ECL")
	}
	for _, st := range ctx.Stations {
		addPoint(st.NatPoint, "STAT")
	}
	if ctx.Profection != nil {
		addPoint(ctx.Profection.Lord, "PROF")
	}
	if ctx.Firdaria != nil {
		addPoint(ctx.Firdaria.MajorLord, "FIRD")
	}

	result := make(map[string]int)
	for point, techs := range points {
		result[point] = len(techs)
	}
	return result
}

// DominantThemes returns the top N most-activated planets across all techniques.
func DominantThemes(ctx *FullContext, topN int) []string {
	scores := ConvergenceScoreByPoint(ctx)
	type kv struct{ k string; v int }
	var sorted []kv
	for k, v := range scores {
		sorted = append(sorted, kv{k, v})
	}
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].v > sorted[i].v {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	var result []string
	for i := 0; i < topN && i < len(sorted); i++ {
		keywords := astromath.PlanetKeywords[sorted[i].k]
		if keywords == "" { keywords = sorted[i].k }
		result = append(result, fmt.Sprintf("%s (%d técnicas) — %s", sorted[i].k, sorted[i].v, keywords))
	}
	return result
}

// MonthScores computes a 0-100 activation score for each month.
func MonthScores(ctx *FullContext) [12]int {
	var scores [12]int

	for _, tr := range ctx.Transits {
		for _, ep := range tr.EpDetails {
			for m := ep.MonthStart; m <= ep.MonthEnd && m <= 12; m++ {
				if m >= 1 { scores[m-1] += 8 }
			}
		}
	}
	for _, ecl := range ctx.Eclipses {
		if ecl.Eclipse.Month >= 1 && ecl.Eclipse.Month <= 12 {
			scores[ecl.Eclipse.Month-1] += 12
		}
	}
	for _, st := range ctx.Stations {
		if st.NatPoint != "" && st.Month >= 1 && st.Month <= 12 {
			scores[st.Month-1] += 15
		}
	}
	for _, et := range ctx.EclipseTriggers {
		if et.Month >= 1 && et.Month <= 12 {
			scores[et.Month-1] += 10
		}
	}
	// Normalize each month to 0-100
	maxScore := 1
	for _, s := range scores {
		if s > maxScore { maxScore = s }
	}
	for i := range scores {
		scores[i] = int(float64(scores[i]) / float64(maxScore) * 100)
	}
	return scores
}

// NuclearMonth represents the month with highest technique convergence.
type NuclearMonth struct {
	Month      int      `json:"month"`       // 1-12
	Score      int      `json:"score"`       // 0-100
	Techniques int      `json:"techniques"`  // count of techniques converging
}

// FindNuclearMonth returns the month with the highest convergence score.
// Injected into brief as "MES NUCLEAR: Mayo (score 92 — 4 técnicas convergen)".
func FindNuclearMonth(monthlyScores [12]int) *NuclearMonth {
	bestMonth := 0
	bestScore := 0
	for i, s := range monthlyScores {
		if s > bestScore {
			bestScore = s
			bestMonth = i + 1
		}
	}
	if bestScore < 30 {
		return nil // no significant convergence
	}

	// Count how many months have >50% of the nuclear month's score
	techCount := 0
	for _, s := range monthlyScores {
		if s >= bestScore/2 {
			techCount++
		}
	}

	return &NuclearMonth{
		Month:      bestMonth,
		Score:      bestScore,
		Techniques: techCount,
	}
}

// BirthTimeStatus returns reliability of birth time and which techniques to trust.
type BirthTimeReliability struct {
	Known        bool     `json:"known"`
	Reliability  string   `json:"reliability"` // "alta", "media", "baja"
	TrustAll     bool     `json:"trust_all"`   // can trust house-dependent techniques
	Description  string   `json:"description"`
}

func AssessBirthTime(chart *FullContext, birthTimeKnown bool) *BirthTimeReliability {
	if birthTimeKnown {
		return &BirthTimeReliability{
			Known: true, Reliability: "alta", TrustAll: true,
			Description: "Hora de nacimiento conocida — todas las técnicas son confiables",
		}
	}
	return &BirthTimeReliability{
		Known: false, Reliability: "baja", TrustAll: false,
		Description: "Hora desconocida — casas, ASC, MC, direcciones primarias y profecciones tienen margen de error. Tránsitos y firdaria son confiables.",
	}
}

// TechniqueVerdict is a typed verdict from a single technique.
type TechniqueVerdict struct {
	Technique string  `json:"technique"`
	Verdict   string  `json:"verdict"`   // "favorable", "challenging", "neutral", "mixed"
	Planet    string  `json:"planet"`    // key planet involved
	Month     int     `json:"month"`     // 0 = year-long
	Strength  float64 `json:"strength"`  // 0-1
	Detail    string  `json:"detail"`
}

// ExtractVerdicts produces one verdict per active technique.
func ExtractVerdicts(ctx *FullContext) []TechniqueVerdict {
	var verdicts []TechniqueVerdict

	// Profection verdict
	if ctx.Profection != nil {
		verdicts = append(verdicts, TechniqueVerdict{
			Technique: "profecciones", Verdict: "neutral",
			Planet: ctx.Profection.Lord, Strength: 0.8,
			Detail: fmt.Sprintf("Casa %d activa, cronócrata %s — %s",
				ctx.Profection.ActiveHouse, ctx.Profection.Lord, ctx.Profection.Theme),
		})
	}

	// Firdaria verdict
	if ctx.Firdaria != nil {
		verdicts = append(verdicts, TechniqueVerdict{
			Technique: "firdaria", Verdict: "neutral",
			Planet: ctx.Firdaria.MajorLord, Strength: 0.7,
			Detail: fmt.Sprintf("Período mayor %s, sub %s", ctx.Firdaria.MajorLord, ctx.Firdaria.SubLord),
		})
	}

	// Transit verdicts (slow planets only)
	for _, tr := range ctx.Transits {
		verdict := "neutral"
		switch tr.Nature {
		case "fácil": verdict = "favorable"
		case "tenso": verdict = "challenging"
		}
		verdicts = append(verdicts, TechniqueVerdict{
			Technique: "tránsito", Verdict: verdict,
			Planet: tr.Transit, Month: tr.Month, Strength: 1.0 - tr.Orb/3.0,
			Detail: fmt.Sprintf("%s %s %s", tr.Transit, tr.Aspect, tr.Natal),
		})
	}

	// SA verdicts
	for _, sa := range ctx.SolarArc {
		verdict := "neutral"
		if sa.Nature == "fácil" { verdict = "favorable" }
		if sa.Nature == "tenso" { verdict = "challenging" }
		verdicts = append(verdicts, TechniqueVerdict{
			Technique: "arco_solar", Verdict: verdict,
			Planet: sa.SAplanet, Strength: 1.0 - sa.Orb/1.5,
			Detail: fmt.Sprintf("SA %s %s %s", sa.SAplanet, sa.Aspect, sa.NatPlanet),
		})
	}

	// Eclipse verdicts
	for _, ecl := range ctx.Eclipses {
		verdicts = append(verdicts, TechniqueVerdict{
			Technique: "eclipse", Verdict: "challenging",
			Planet: ecl.NatPoint, Month: ecl.Eclipse.Month, Strength: 0.9,
			Detail: fmt.Sprintf("Eclipse %s → %s", ecl.Eclipse.Type, ecl.NatPoint),
		})
	}

	return verdicts
}

// Contradiction records when two techniques give opposing verdicts on the same theme.
type Contradiction struct {
	TechA   string `json:"tech_a"`
	TechB   string `json:"tech_b"`
	VerdictA string `json:"verdict_a"`
	VerdictB string `json:"verdict_b"`
	Planet  string `json:"planet"`
	Resolution string `json:"resolution"`
}

// ResolveContradictions finds and resolves opposing verdicts.
func ResolveContradictions(verdicts []TechniqueVerdict) []Contradiction {
	var contradictions []Contradiction

	// Group verdicts by planet
	byPlanet := make(map[string][]TechniqueVerdict)
	for _, v := range verdicts {
		if v.Planet != "" {
			byPlanet[v.Planet] = append(byPlanet[v.Planet], v)
		}
	}

	for planet, vs := range byPlanet {
		hasFavorable := false
		hasChallenging := false
		var favTech, chalTech string
		for _, v := range vs {
			if v.Verdict == "favorable" { hasFavorable = true; favTech = v.Technique }
			if v.Verdict == "challenging" { hasChallenging = true; chalTech = v.Technique }
		}
		if hasFavorable && hasChallenging {
			resolution := fmt.Sprintf("%s da señal mixta: oportunidad (%s) con tensión (%s). "+
				"La dirección depende de la acción consciente.", planet, favTech, chalTech)
			contradictions = append(contradictions, Contradiction{
				TechA: favTech, TechB: chalTech,
				VerdictA: "favorable", VerdictB: "challenging",
				Planet: planet, Resolution: resolution,
			})
		}
	}

	return contradictions
}

// SynthesisBrief builds a synthesis section from verdicts and contradictions.
func SynthesisBrief(verdicts []TechniqueVerdict, contradictions []Contradiction, monthScores [12]int) string {
	var b strings.Builder

	b.WriteString("## SÍNTESIS\n\n")

	// Count verdicts by type
	fav, chal, neut := 0, 0, 0
	for _, v := range verdicts {
		switch v.Verdict {
		case "favorable": fav++
		case "challenging": chal++
		default: neut++
		}
	}
	total := fav + chal + neut
	if total > 0 {
		b.WriteString(fmt.Sprintf("Balance: %d favorables, %d desafiantes, %d neutros\n", fav, chal, neut))
		if fav > chal*2 {
			b.WriteString("**Tendencia general: FAVORABLE** — mayoría de técnicas apoyan\n")
		} else if chal > fav*2 {
			b.WriteString("**Tendencia general: DESAFIANTE** — requiere gestión activa\n")
		} else {
			b.WriteString("**Tendencia general: MIXTA** — oportunidades con desafíos\n")
		}
	}

	// Contradictions
	if len(contradictions) > 0 {
		b.WriteString("\n**Señales contradictorias:**\n")
		for _, c := range contradictions {
			b.WriteString(fmt.Sprintf("- %s\n", c.Resolution))
		}
	}

	// Best/worst months
	bestMonth, worstMonth := 0, 0
	bestScore, worstScore := 0, 101
	for i, s := range monthScores {
		if s > bestScore { bestScore = s; bestMonth = i + 1 }
		if s < worstScore { worstScore = s; worstMonth = i + 1 }
	}
	if bestMonth > 0 {
		b.WriteString(fmt.Sprintf("\n**Mes más activo:** %d (score %d/100)\n", bestMonth, bestScore))
		b.WriteString(fmt.Sprintf("**Mes más tranquilo:** %d (score %d/100)\n", worstMonth, worstScore))
	}

	return b.String()
}

// FilterTopN trims activation lists to the top N entries per technique,
// sorted by tightest orb. Prevents briefs with 50+ activations where the LLM
// loses focus. Applied before BuildBrief() in the builder pipeline.
func FilterTopN(ctx *FullContext, maxPerTechnique int) {
	if maxPerTechnique <= 0 {
		maxPerTechnique = 10
	}

	// Directions: sort by orb ascending, keep top N
	if len(ctx.Directions) > maxPerTechnique {
		sortDirectionsByOrb(ctx.Directions)
		ctx.Directions = ctx.Directions[:maxPerTechnique]
	}

	// Solar Arc: sort by orb ascending, keep top N
	if len(ctx.SolarArc) > maxPerTechnique {
		sortSolarArcByOrb(ctx.SolarArc)
		ctx.SolarArc = ctx.SolarArc[:maxPerTechnique]
	}

	// Transits: keep top N by tightest orb
	if len(ctx.Transits) > maxPerTechnique {
		sortTransitsByOrb(ctx.Transits)
		ctx.Transits = ctx.Transits[:maxPerTechnique]
	}

	// Fast transits: keep top N
	if len(ctx.FastTransits) > maxPerTechnique {
		sortFastTransitsByOrb(ctx.FastTransits)
		ctx.FastTransits = ctx.FastTransits[:maxPerTechnique]
	}

	// Eclipses: keep top N
	if len(ctx.Eclipses) > maxPerTechnique {
		sortEclipsesByOrb(ctx.Eclipses)
		ctx.Eclipses = ctx.Eclipses[:maxPerTechnique]
	}

	// Eclipse triggers: keep top N
	if len(ctx.EclipseTriggers) > maxPerTechnique {
		ctx.EclipseTriggers = ctx.EclipseTriggers[:maxPerTechnique]
	}
}

// Sort helpers — sort by orb ascending (tightest first).

func sortDirectionsByOrb(ds []technique.PrimaryDirection) {
	for i := 0; i < len(ds); i++ {
		for j := i + 1; j < len(ds); j++ {
			if ds[j].OrbDeg < ds[i].OrbDeg {
				ds[i], ds[j] = ds[j], ds[i]
			}
		}
	}
}

func sortSolarArcByOrb(sas []technique.SolarArcResult) {
	for i := 0; i < len(sas); i++ {
		for j := i + 1; j < len(sas); j++ {
			if sas[j].Orb < sas[i].Orb {
				sas[i], sas[j] = sas[j], sas[i]
			}
		}
	}
}

func sortTransitsByOrb(trs []technique.TransitActivation) {
	for i := 0; i < len(trs); i++ {
		for j := i + 1; j < len(trs); j++ {
			if trs[j].Orb < trs[i].Orb {
				trs[i], trs[j] = trs[j], trs[i]
			}
		}
	}
}

func sortFastTransitsByOrb(fts []technique.FastTransitActivation) {
	for i := 0; i < len(fts); i++ {
		for j := i + 1; j < len(fts); j++ {
			if fts[j].Orb < fts[i].Orb {
				fts[i], fts[j] = fts[j], fts[i]
			}
		}
	}
}

func sortEclipsesByOrb(ecls []technique.EclipseActivation) {
	for i := 0; i < len(ecls); i++ {
		for j := i + 1; j < len(ecls); j++ {
			if ecls[j].Orb < ecls[i].Orb {
				ecls[i], ecls[j] = ecls[j], ecls[i]
			}
		}
	}
}
