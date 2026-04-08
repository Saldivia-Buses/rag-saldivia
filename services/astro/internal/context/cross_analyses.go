package context

import (
	"fmt"
	"strings"

	"github.com/Camionerou/rag-saldivia/services/astro/internal/astromath"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/ephemeris"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/natal"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/technique"
)

// RSLRCrossing represents a Solar Return × Lunar Return monthly crossing.
type RSLRCrossing struct {
	Month     int    `json:"month"`
	LRHouseInRS int  `json:"lr_house_in_rs"` // LR ASC falls in which RS house
	Aspects   []string `json:"aspects"`
}

// CalcRSLRCrossings computes monthly Solar Return × Lunar Return crossings.
// Shows which RS house the LR ASC falls in each month.
func CalcRSLRCrossings(sr *technique.SolarReturn, lrs []technique.LunarReturn) []RSLRCrossing {
	if sr == nil { return nil }
	var crossings []RSLRCrossing

	for _, lr := range lrs {
		// Compute LR chart ASC (simplified: use the Moon's position as proxy for LR energy)
		// Full implementation would build a chart at the LR moment
		lrLon := 0.0
		// The LR JD tells us the month
		_, m, _, _ := ephemeris.RevJul(lr.JD)

		// LR "ASC" approximation: the Moon's natal position is the LR focus
		// In a full implementation, we'd build houses at LR moment
		// For now, use the LR JD to estimate where the Moon falls in the SR chart
		moonPos, err := ephemeris.CalcPlanet(lr.JD, ephemeris.Moon, ephemeris.FlagSwieph|ephemeris.FlagSpeed)
		if err != nil { continue }
		lrLon = moonPos.Lon

		// Which SR house does the LR Moon fall in?
		lrHouse := astromath.HouseForLon(lrLon, sr.Cusps)

		// Check aspects between LR Moon and SR planets
		var aspects []string
		for name, pos := range sr.Planets {
			asp := astromath.FindAspect(lrLon, pos.Lon, 5.0)
			if asp != nil {
				aspects = append(aspects, fmt.Sprintf("LR Luna %s SR %s", asp.Name, name))
			}
		}

		crossings = append(crossings, RSLRCrossing{
			Month: m, LRHouseInRS: lrHouse, Aspects: aspects,
		})
	}

	return crossings
}

// PrenatalEclipseTransits checks if SA, transits, or PD activate the prenatal eclipse degree.
type PrenatalEclipseActivation struct {
	Technique string  `json:"technique"` // "SA", "transit", "PD"
	Planet    string  `json:"planet"`
	Aspect    string  `json:"aspect"`
	Orb       float64 `json:"orb"`
	Month     int     `json:"month"`
	EclType   string  `json:"ecl_type"` // "solar" or "lunar"
}

func CalcPrenatalEclipseTransits(
	chart *natal.Chart,
	prenatal *technique.PrenatalEclipseResult,
	solarArcs []technique.SolarArcResult,
	transits []technique.TransitActivation,
	directions []technique.PrimaryDirection,
	year int,
) []PrenatalEclipseActivation {
	if prenatal == nil { return nil }
	var activations []PrenatalEclipseActivation

	eclipsePoints := make(map[string]float64) // "solar"/"lunar" → lon
	if prenatal.Solar != nil { eclipsePoints["solar"] = prenatal.Solar.Lon }
	if prenatal.Lunar != nil { eclipsePoints["lunar"] = prenatal.Lunar.Lon }

	for eclType, eclLon := range eclipsePoints {
		// SA activations
		for _, sa := range solarArcs {
			asp := astromath.FindAspect(sa.SALon, eclLon, 1.5)
			if asp != nil {
				activations = append(activations, PrenatalEclipseActivation{
					Technique: "SA", Planet: sa.SAplanet, Aspect: asp.Name,
					Orb: asp.Orb, EclType: eclType,
				})
			}
		}

		// Transit activations
		for _, tr := range transits {
			// Check if transit planet hits the prenatal eclipse degree
			flags := ephemeris.FlagSwieph | ephemeris.FlagSpeed
			for m := 1; m <= 12; m++ {
				jd := ephemeris.JulDay(year, m, 15, 12.0)
				for _, sp := range []struct{ name string; id int }{
					{"Saturno", ephemeris.Saturn}, {"Júpiter", ephemeris.Jupiter},
					{"Marte", ephemeris.Mars},
				} {
					if sp.name != tr.Transit { continue }
					pos, err := ephemeris.CalcPlanet(jd, sp.id, flags)
					if err != nil { continue }
					asp := astromath.FindAspect(pos.Lon, eclLon, 3.0)
					if asp != nil {
						activations = append(activations, PrenatalEclipseActivation{
							Technique: "transit", Planet: sp.name, Aspect: asp.Name,
							Orb: asp.Orb, Month: m, EclType: eclType,
						})
					}
				}
				break // only check once per transit planet
			}
		}
	}

	return activations
}

// DivisorResult holds the current Divisor (term lord of the directed ASC).
type DivisorResult struct {
	Lord           string  `json:"lord"`
	RemainingYears float64 `json:"remaining_years"`
	NextLord       string  `json:"next_lord"`
	Description    string  `json:"description"`
}

// CalcDivisor computes the Divisor — the term lord of the directed ASC.
// ASC directed by age × Naibod rate through the terms.
func CalcDivisor(chart *natal.Chart, age float64) *DivisorResult {
	arc := age * astromath.NaibodRate
	directedASC := astromath.Normalize360(chart.ASC + arc)
	currentLord := astromath.TermLord(directedASC)

	// Find how many degrees until the next term boundary
	signIdx := int(directedASC / 30)
	degInSign := directedASC - float64(signIdx*30)
	terms := astromath.PtolemaicTerms[signIdx%12]

	remainingDeg := 0.0
	nextLord := ""
	for i, t := range terms {
		if degInSign < t.UpperDeg {
			remainingDeg = t.UpperDeg - degInSign
			if i+1 < len(terms) {
				nextLord = terms[i+1].Lord
			} else {
				// Next sign's first term
				nextTerms := astromath.PtolemaicTerms[(signIdx+1)%12]
				if len(nextTerms) > 0 { nextLord = nextTerms[0].Lord }
			}
			break
		}
	}
	remainingYears := remainingDeg / astromath.NaibodRate

	return &DivisorResult{
		Lord:           currentLord,
		RemainingYears: remainingYears,
		NextLord:       nextLord,
		Description:    fmt.Sprintf("Divisor actual: %s (%.1f años restantes, próximo: %s)", currentLord, remainingYears, nextLord),
	}
}

// TriplicityLordsResult holds the life division by triplicity lords.
type TriplicityLordsResult struct {
	SectLight  string           `json:"sect_light"`
	Element    string           `json:"element"`
	Periods    []TriplicityPeriod `json:"periods"`
	Active     string           `json:"active_lord"`
}

type TriplicityPeriod struct {
	Lord    string  `json:"lord"`
	Phase   string  `json:"phase"` // "primera vida", "segunda vida", "tercera vida"
	AgeFrom float64 `json:"age_from"`
	AgeTo   float64 `json:"age_to"`
	Active  bool    `json:"active"`
}

// CalcTriplicityLords divides life into 3 phases based on sect luminary's triplicity lords.
func CalcTriplicityLords(chart *natal.Chart, age float64) *TriplicityLordsResult {
	sectLight := "Sol"
	if !chart.Diurnal { sectLight = "Luna" }

	sectPos, ok := chart.Planets[sectLight]
	if !ok { return nil }

	signIdx := astromath.SignIndex(sectPos.Lon)
	elems := [12]string{"fire", "earth", "air", "water", "fire", "earth", "air", "water", "fire", "earth", "air", "water"}
	elem := elems[signIdx]
	trip, ok := astromath.Triplicity[elem]
	if !ok { return nil }

	dayLord := trip.Day
	nightLord := trip.Night
	// Third lord: participating triplicity lord (varies by tradition)
	// Simplified: use the remaining classical planet associated with the element
	thirdLord := "Saturno" // default participating lord

	// Life division: roughly thirds
	periods := []TriplicityPeriod{
		{Lord: dayLord, Phase: "primera vida", AgeFrom: 0, AgeTo: 30, Active: age < 30},
		{Lord: nightLord, Phase: "segunda vida", AgeFrom: 30, AgeTo: 60, Active: age >= 30 && age < 60},
		{Lord: thirdLord, Phase: "tercera vida", AgeFrom: 60, AgeTo: 90, Active: age >= 60},
	}

	active := dayLord
	if age >= 30 && age < 60 { active = nightLord }
	if age >= 60 { active = thirdLord }

	return &TriplicityLordsResult{
		SectLight: sectLight, Element: elem, Periods: periods, Active: active,
	}
}

// ChronocratorFirdariaCross detects when profection lord = firdaria lord (double convergence).
type ChronocratorCross struct {
	Match       bool   `json:"match"`
	Planet      string `json:"planet"`
	Description string `json:"description"`
}

func CalcChronocratorFirdariaCross(profection *technique.Profection, firdaria *technique.Firdaria) *ChronocratorCross {
	if profection == nil || firdaria == nil {
		return &ChronocratorCross{Match: false}
	}
	if profection.Lord == firdaria.MajorLord {
		return &ChronocratorCross{
			Match: true, Planet: profection.Lord,
			Description: fmt.Sprintf("DOBLE CONVERGENCIA: %s es cronócrata Y señor firdaria mayor — tema amplificado", profection.Lord),
		}
	}
	if profection.Lord == firdaria.SubLord {
		return &ChronocratorCross{
			Match: true, Planet: profection.Lord,
			Description: fmt.Sprintf("Convergencia parcial: %s es cronócrata Y sub-señor firdaria", profection.Lord),
		}
	}
	return &ChronocratorCross{Match: false}
}

// TablaRow represents one row in the multi-entity comparison table.
type TablaRow struct {
	Entity string            `json:"entity"` // contact name
	Year   int               `json:"year"`
	Cells  map[string]string `json:"cells"`  // technique → summary
	Score  int               `json:"score"`
}

// BuildTabla generates a multi-entity × multi-year comparison table.
// Each cell contains a brief summary of the technique for that entity+year.
func BuildTabla(entities []TablaEntity, yearStart, yearEnd int) []TablaRow {
	var rows []TablaRow

	for _, e := range entities {
		for year := yearStart; year <= yearEnd; year++ {
			cells := make(map[string]string)

			// Profection
			if e.Profection != nil {
				cells["profeccion"] = fmt.Sprintf("Casa %d (%s)", e.Profection.ActiveHouse, e.Profection.Lord)
			}

			// Firdaria
			if e.Firdaria != nil {
				cells["firdaria"] = fmt.Sprintf("%s/%s", e.Firdaria.MajorLord, e.Firdaria.SubLord)
			}

			// SA count
			cells["SA"] = fmt.Sprintf("%d activaciones", len(e.SolarArcs))

			// Transit summary
			tenso, facil := 0, 0
			for _, tr := range e.Transits {
				if tr.Nature == "tenso" { tenso++ } else { facil++ }
			}
			cells["transitos"] = fmt.Sprintf("%d fácil / %d tenso", facil, tenso)

			// Eclipse count
			cells["eclipses"] = fmt.Sprintf("%d", len(e.Eclipses))

			// Score
			score := len(e.SolarArcs)*3 + len(e.Transits)*2 + len(e.Eclipses)*4
			if score > 100 { score = 100 }

			rows = append(rows, TablaRow{
				Entity: e.Name, Year: year, Cells: cells, Score: score,
			})
		}
	}

	return rows
}

// TablaEntity holds pre-computed data for one entity in the table.
type TablaEntity struct {
	Name       string
	Profection *technique.Profection
	Firdaria   *technique.Firdaria
	SolarArcs  []technique.SolarArcResult
	Transits   []technique.TransitActivation
	Eclipses   []technique.EclipseActivation
}

// FormatCrossAnalyses generates text for all cross-technique analyses.
func FormatCrossAnalyses(
	rsLR []RSLRCrossing,
	prenatalTransits []PrenatalEclipseActivation,
	divisor *DivisorResult,
	triplicity *TriplicityLordsResult,
	chronoCross *ChronocratorCross,
) string {
	var b strings.Builder

	if divisor != nil {
		b.WriteString(fmt.Sprintf("## DIVISOR\n%s\n\n", divisor.Description))
	}

	if triplicity != nil {
		b.WriteString(fmt.Sprintf("## SEÑORES DE TRIPLICIDAD\nLuz de secta: %s (%s)\nSeñor activo: %s\n",
			triplicity.SectLight, triplicity.Element, triplicity.Active))
		for _, p := range triplicity.Periods {
			marker := ""
			if p.Active { marker = " ← ACTIVO" }
			b.WriteString(fmt.Sprintf("  %s: %s (%.0f-%.0f)%s\n", p.Phase, p.Lord, p.AgeFrom, p.AgeTo, marker))
		}
		b.WriteString("\n")
	}

	if chronoCross != nil && chronoCross.Match {
		b.WriteString(fmt.Sprintf("## CONVERGENCIA CRONÓCRATA-FIRDARIA\n⚠ %s\n\n", chronoCross.Description))
	}

	if len(prenatalTransits) > 0 {
		b.WriteString("## ACTIVACIONES ECLIPSE PRENATAL\n")
		for _, a := range prenatalTransits {
			b.WriteString(fmt.Sprintf("- %s %s %s eclipse %s (orbe %.1f°)\n",
				a.Technique, a.Planet, a.Aspect, a.EclType, a.Orb))
		}
		b.WriteString("\n")
	}

	if len(rsLR) > 0 {
		b.WriteString("## CRUCES RS × RL\n")
		for _, c := range rsLR {
			b.WriteString(fmt.Sprintf("  Mes %d: RL en casa RS %d", c.Month, c.LRHouseInRS))
			if len(c.Aspects) > 0 {
				b.WriteString(fmt.Sprintf(" — %s", strings.Join(c.Aspects, ", ")))
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	return b.String()
}
