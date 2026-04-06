package technique

import (
	"github.com/Camionerou/rag-saldivia/services/astro/internal/astromath"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/ephemeris"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/natal"
)

// Eclipse holds data about a solar or lunar eclipse.
type Eclipse struct {
	Type      string  `json:"type"`      // "solar" or "lunar"
	SubType   string  `json:"sub_type"`  // "total", "annular", "partial"
	JD        float64 `json:"jd"`
	Lon       float64 `json:"lon"`       // ecliptic longitude of eclipse
	Month     int     `json:"month"`
	Day       int     `json:"day"`
}

// EclipseActivation records when an eclipse contacts a natal point.
type EclipseActivation struct {
	Eclipse   Eclipse `json:"eclipse"`
	NatPoint  string  `json:"natal_point"`
	NatLon    float64 `json:"natal_lon"`
	Orb       float64 `json:"orb"`
	Aspect    string  `json:"aspect"`
}

const eclipseNatalOrb = 3.0 // degrees

// FindEclipses finds all solar and lunar eclipses in a given year.
func FindEclipses(targetYear int) ([]Eclipse, error) {
	jdStart := ephemeris.JulDay(targetYear, 1, 1, 0.0)
	jdEnd := ephemeris.JulDay(targetYear+1, 1, 1, 0.0)
	flags := ephemeris.FlagSwieph

	var eclipses []Eclipse

	// Solar eclipses
	jd := jdStart
	for jd < jdEnd {
		tret, err := ephemeris.SolEclipseWhenGlob(jd, flags, 0)
		if err != nil || tret[0] >= jdEnd {
			break
		}
		eclJD := tret[0]
		// Get Sun longitude at eclipse time
		sunPos, err := ephemeris.CalcPlanet(eclJD, ephemeris.Sun, flags|ephemeris.FlagSpeed)
		if err != nil {
			jd = eclJD + 30
			continue
		}
		_, m, d, _ := ephemeris.RevJul(eclJD)
		eclipses = append(eclipses, Eclipse{
			Type:    "solar",
			SubType: eclipseSubType(tret),
			JD:      eclJD,
			Lon:     sunPos.Lon,
			Month:   m,
			Day:     d,
		})
		jd = eclJD + 30
	}

	// Lunar eclipses
	jd = jdStart
	for jd < jdEnd {
		tret, err := ephemeris.LunEclipseWhen(jd, flags, 0)
		if err != nil || tret[0] >= jdEnd {
			break
		}
		eclJD := tret[0]
		// Lunar eclipse: Moon opposite Sun, so eclipse lon = Moon lon
		moonPos, err := ephemeris.CalcPlanet(eclJD, ephemeris.Moon, flags|ephemeris.FlagSpeed)
		if err != nil {
			jd = eclJD + 30
			continue
		}
		_, m, d, _ := ephemeris.RevJul(eclJD)
		eclipses = append(eclipses, Eclipse{
			Type:    "lunar",
			SubType: eclipseSubType(tret),
			JD:      eclJD,
			Lon:     moonPos.Lon,
			Month:   m,
			Day:     d,
		})
		jd = eclJD + 30
	}

	return eclipses, nil
}

// FindEclipseActivations checks which eclipses aspect natal points.
func FindEclipseActivations(chart *natal.Chart, targetYear int) ([]EclipseActivation, error) {
	eclipses, err := FindEclipses(targetYear)
	if err != nil {
		return nil, err
	}

	natalPoints := make(map[string]float64)
	for name, pos := range chart.Planets {
		natalPoints[name] = pos.Lon
	}
	natalPoints["ASC"] = chart.ASC
	natalPoints["MC"] = chart.MC

	var activations []EclipseActivation
	for _, ecl := range eclipses {
		for natName, natLon := range natalPoints {
			asp := astromath.FindAspect(ecl.Lon, natLon, eclipseNatalOrb)
			if asp != nil {
				activations = append(activations, EclipseActivation{
					Eclipse:  ecl,
					NatPoint: natName,
					NatLon:   natLon,
					Orb:      asp.Orb,
					Aspect:   asp.Name,
				})
			}
		}
	}

	return activations, nil
}

func eclipseSubType(tret []float64) string {
	// tret[0] = max eclipse JD, type info from return flags
	// Simplified: check if total, annular, or partial based on tret values
	if len(tret) > 4 && tret[4] > 0 {
		return "total"
	}
	if len(tret) > 5 && tret[5] > 0 {
		return "annular"
	}
	return "partial"
}
