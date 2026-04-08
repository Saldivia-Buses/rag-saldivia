package context

import (
	"fmt"
	"strings"

	"github.com/Camionerou/rag-saldivia/services/astro/internal/astromath"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/ephemeris"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/natal"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/technique"
)

// AntisciaContext generates text for antiscia natal aspects and transits.
func AntisciaContext(chart *natal.Chart, year int) string {
	var b strings.Builder
	b.WriteString("## ANTISCIA\n\n")

	// Natal antiscia: check all planet pairs for antiscion conjunctions
	names := []string{"Sol", "Luna", "Mercurio", "Venus", "Marte", "Júpiter", "Saturno"}
	found := 0
	for i := 0; i < len(names); i++ {
		posA, okA := chart.Planets[names[i]]
		if !okA { continue }
		antA := astromath.Antiscion(posA.Lon)
		contraA := astromath.ContraAntiscion(posA.Lon)
		for j := i + 1; j < len(names); j++ {
			posB, okB := chart.Planets[names[j]]
			if !okB { continue }
			if astromath.AngDiff(antA, posB.Lon) < 1.5 {
				b.WriteString(fmt.Sprintf("- Antiscia: %s ↔ %s (orbe %.1f°)\n",
					names[i], names[j], astromath.AngDiff(antA, posB.Lon)))
				found++
			}
			if astromath.AngDiff(contraA, posB.Lon) < 1.5 {
				b.WriteString(fmt.Sprintf("- Contra-antiscia: %s ↔ %s (orbe %.1f°)\n",
					names[i], names[j], astromath.AngDiff(contraA, posB.Lon)))
				found++
			}
		}
	}
	if found == 0 {
		b.WriteString("Sin antiscias natales dentro del orbe.\n")
	}
	b.WriteString("\n")
	return b.String()
}

// FixedStarsTransitContext checks which fixed stars are activated by slow transits during the year.
func FixedStarsTransitContext(year int) string {
	var b strings.Builder
	b.WriteString("## ESTRELLAS FIJAS EN TRÁNSITO\n\n")

	flags := ephemeris.FlagSwieph | ephemeris.FlagSpeed
	jdMid := ephemeris.JulDay(year, 7, 1, 12.0)
	slowPlanets := []struct{ name string; id int }{
		{"Júpiter", ephemeris.Jupiter}, {"Saturno", ephemeris.Saturn},
		{"Urano", ephemeris.Uranus}, {"Neptuno", ephemeris.Neptune}, {"Plutón", ephemeris.Pluto},
	}

	found := 0
	for _, sp := range slowPlanets {
		pos, err := ephemeris.CalcPlanet(jdMid, sp.id, flags)
		if err != nil { continue }
		for _, star := range astromath.MajorFixedStars {
			starLon, err := ephemeris.FixstarUT(star.SweName, jdMid, flags)
			if err != nil { continue }
			orb := astromath.AngDiff(pos.Lon, starLon)
			if orb <= 1.5 {
				b.WriteString(fmt.Sprintf("- %s conjunción %s (orbe %.1f°, naturaleza %s)\n",
					sp.name, star.Name, orb, star.Nature))
				found++
			}
		}
	}
	if found == 0 {
		b.WriteString("Sin conjunciones de planetas lentos a estrellas fijas.\n")
	}
	b.WriteString("\n")
	return b.String()
}

// CazimiCombustionTransitContext tracks cazimi/combustion events during the year.
func CazimiCombustionTransitContext(year int) string {
	var b strings.Builder
	b.WriteString("## CAZIMI Y COMBUSTIÓN EN TRÁNSITO\n\n")

	flags := ephemeris.FlagSwieph | ephemeris.FlagSpeed
	planets := []struct{ name string; id int }{
		{"Mercurio", ephemeris.Mercury}, {"Venus", ephemeris.Venus},
		{"Marte", ephemeris.Mars},
	}

	found := 0
	for month := 1; month <= 12; month++ {
		jd := ephemeris.JulDay(year, month, 15, 12.0)
		sunPos, err := ephemeris.CalcPlanet(jd, ephemeris.Sun, flags)
		if err != nil { continue }
		for _, p := range planets {
			pos, err := ephemeris.CalcPlanet(jd, p.id, flags)
			if err != nil { continue }
			status := astromath.CombustionStatus(pos.Lon, sunPos.Lon)
			if status != "" {
				b.WriteString(fmt.Sprintf("- %s %s (mes %d)\n", p.name, status, month))
				found++
			}
		}
	}
	if found == 0 {
		b.WriteString("Sin eventos de cazimi/combustión significativos.\n")
	}
	b.WriteString("\n")
	return b.String()
}

// DavisonTransitsContext overlays transits onto a Davison relationship chart.
func DavisonTransitsContext(davison *technique.DavisonResult, year int) string {
	if davison == nil || davison.Chart == nil { return "" }
	transits := technique.CalcTransits(davison.Chart, year)
	if len(transits) == 0 { return "" }

	var b strings.Builder
	b.WriteString("## TRÁNSITOS SOBRE CARTA DAVISON\n\n")
	for _, tr := range transits {
		b.WriteString(fmt.Sprintf("- %s %s %s (orbe %.1f°, %s)\n",
			tr.Transit, tr.Aspect, tr.Natal, tr.Orb, tr.Nature))
	}
	b.WriteString("\n")
	return b.String()
}

// VOCPeriod represents a Void-of-Course Moon period.
type VOCPeriodInfo struct {
	StartMonth int     `json:"start_month"`
	StartDay   int     `json:"start_day"`
	EndMonth   int     `json:"end_month"`
	EndDay     int     `json:"end_day"`
	DurationH  float64 `json:"duration_hours"`
	FromSign   string  `json:"from_sign"`
	ToSign     string  `json:"to_sign"`
}

// CalcVOCPeriods finds Void-of-Course Moon periods for a month.
// VOC = Moon makes no major aspects before leaving its current sign.
func CalcVOCPeriods(year, month int) []VOCPeriodInfo {
	flags := ephemeris.FlagSwieph | ephemeris.FlagSpeed
	jdStart := ephemeris.JulDay(year, month, 1, 0.0)
	jdEnd := ephemeris.JulDay(year, month+1, 0, 0.0)
	if month == 12 { jdEnd = ephemeris.JulDay(year+1, 1, 0, 0.0) }

	var periods []VOCPeriodInfo
	prevSign := -1

	for jd := jdStart; jd < jdEnd; jd += 0.25 { // every 6 hours
		moonPos, err := ephemeris.CalcPlanet(jd, ephemeris.Moon, flags)
		if err != nil { continue }
		curSign := astromath.SignIndex(moonPos.Lon)

		if prevSign >= 0 && curSign != prevSign {
			// Sign change detected — the VOC period ended here
			// (Simplified: mark the sign transition point)
			_, m1, d1, _ := ephemeris.RevJul(jd - 0.25)
			_, m2, d2, _ := ephemeris.RevJul(jd)
			periods = append(periods, VOCPeriodInfo{
				StartMonth: m1, StartDay: d1,
				EndMonth: m2, EndDay: d2,
				DurationH: 6, // simplified — real VOC duration requires last-aspect detection
				FromSign: astromath.Signs[prevSign],
				ToSign: astromath.Signs[curSign],
			})
		}
		prevSign = curSign
	}

	return periods
}
