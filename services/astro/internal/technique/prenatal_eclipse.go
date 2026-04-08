package technique

import (
	"github.com/Camionerou/rag-saldivia/services/astro/internal/astromath"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/ephemeris"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/natal"
)

// PrenatalEclipse holds the last eclipse before birth.
// In Hellenistic astrology, the prenatal eclipse degree is a sensitive point for the entire life.
type PrenatalEclipse struct {
	Type    string  `json:"type"`    // "solar" or "lunar"
	JD      float64 `json:"jd"`
	Lon     float64 `json:"lon"`
	Sign    string  `json:"sign"`
	Pos     string  `json:"pos"`
	Year    int     `json:"year"`
	Month   int     `json:"month"`
	Day     int     `json:"day"`
}

// PrenatalEclipseResult holds both solar and lunar prenatal eclipses.
type PrenatalEclipseResult struct {
	Solar *PrenatalEclipse `json:"solar"`
	Lunar *PrenatalEclipse `json:"lunar"`
}

// CalcPrenatalEclipses finds the last solar and lunar eclipses before birth.
func CalcPrenatalEclipses(chart *natal.Chart) (*PrenatalEclipseResult, error) {
	flags := ephemeris.FlagSwieph
	result := &PrenatalEclipseResult{}

	// Search backwards for last solar eclipse before birth
	jd := chart.JD
	for i := 0; i < 20; i++ { // max 20 iterations (~10 years back)
		// Search backwards: start from jd - 30 days
		searchJD := jd - 30
		_, tret, err := ephemeris.SolEclipseWhenGlob(searchJD, flags, 0)
		if err != nil {
			break
		}
		eclJD := tret[0]
		if eclJD >= chart.JD {
			jd = searchJD
			continue
		}
		// Found a solar eclipse before birth
		sunPos, err := ephemeris.CalcPlanet(eclJD, ephemeris.Sun, flags|ephemeris.FlagSpeed)
		if err != nil {
			break
		}
		y, m, d, _ := ephemeris.RevJul(eclJD)
		result.Solar = &PrenatalEclipse{
			Type:  "solar",
			JD:    eclJD,
			Lon:   sunPos.Lon,
			Sign:  astromath.SignName(sunPos.Lon),
			Pos:   astromath.PosToStr(sunPos.Lon),
			Year:  y,
			Month: m,
			Day:   d,
		}
		break
	}

	// Search backwards for last lunar eclipse before birth
	jd = chart.JD
	for i := 0; i < 20; i++ {
		searchJD := jd - 30
		_, tret, err := ephemeris.LunEclipseWhen(searchJD, flags, 0)
		if err != nil {
			break
		}
		eclJD := tret[0]
		if eclJD >= chart.JD {
			jd = searchJD
			continue
		}
		moonPos, err := ephemeris.CalcPlanet(eclJD, ephemeris.Moon, flags|ephemeris.FlagSpeed)
		if err != nil {
			break
		}
		y, m, d, _ := ephemeris.RevJul(eclJD)
		result.Lunar = &PrenatalEclipse{
			Type:  "lunar",
			JD:    eclJD,
			Lon:   moonPos.Lon,
			Sign:  astromath.SignName(moonPos.Lon),
			Pos:   astromath.PosToStr(moonPos.Lon),
			Year:  y,
			Month: m,
			Day:   d,
		}
		break
	}

	return result, nil
}
