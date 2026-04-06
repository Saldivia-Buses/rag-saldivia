package technique

import (
	"github.com/Camionerou/rag-saldivia/services/astro/internal/astromath"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/ephemeris"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/natal"
)

// ProgressedPosition holds a progressed planet's position and ingress info.
type ProgressedPosition struct {
	Name         string  `json:"name"`
	Lon          float64 `json:"lon"`
	Sign         string  `json:"sign"`
	Speed        float64 `json:"speed"`
	Retro        bool    `json:"retrograde"`
	House        int     `json:"house"`
	SignIngress  bool    `json:"sign_ingress,omitempty"`
	HouseIngress bool    `json:"house_ingress,omitempty"`
	PrevSign     string  `json:"prev_sign,omitempty"`
	PrevHouse    int     `json:"prev_house,omitempty"`
}

// ProgressionsResult holds all progressed positions for a target year.
type ProgressionsResult struct {
	Year         int                  `json:"year"`
	ProgressedJD float64             `json:"progressed_jd"`
	AgeYears     float64             `json:"age_years"`
	Positions    []ProgressedPosition `json:"positions"`
}

// progressedPlanetOrder ensures deterministic iteration.
var progressedPlanetOrder = []struct {
	Name string
	ID   int
}{
	{"Sol", ephemeris.Sun}, {"Luna", ephemeris.Moon},
	{"Mercurio", ephemeris.Mercury}, {"Venus", ephemeris.Venus},
	{"Marte", ephemeris.Mars}, {"Júpiter", ephemeris.Jupiter},
	{"Saturno", ephemeris.Saturn},
}

// CalcProgressions calculates Secondary Progressions for a target year (mid-year anchor).
// Uses topocentric positions with CalcMu for SetTopo atomicity.
func CalcProgressions(chart *natal.Chart, targetYear int) (*ProgressionsResult, error) {
	jdMid := ephemeris.JulDay(targetYear, 7, 1, 12.0)
	result, err := CalcProgressionsForJD(chart, jdMid)
	if err != nil {
		return nil, err
	}
	result.Year = targetYear
	return result, nil
}

// CalcProgressionsForJD calculates progressions for a precise Julian Day.
func CalcProgressionsForJD(chart *natal.Chart, jdTarget float64) (*ProgressionsResult, error) {
	ageYears := (jdTarget - chart.JD) / 365.25
	progressedJD := chart.JD + ageYears
	prevProgressedJD := chart.JD + (ageYears - 1.0)

	ephemeris.CalcMu.Lock()
	defer ephemeris.CalcMu.Unlock()

	ephemeris.SetTopo(chart.Lon, chart.Lat, chart.Alt)

	flag := ephemeris.FlagSwieph | ephemeris.FlagTopoctr | ephemeris.FlagSpeed
	positions := make([]ProgressedPosition, 0, len(progressedPlanetOrder))

	for _, p := range progressedPlanetOrder {
		pos, err := ephemeris.CalcPlanet(progressedJD, p.ID, flag)
		if err != nil {
			if p.Name == "Sol" || p.Name == "Luna" {
				return nil, err
			}
			continue
		}

		pp := ProgressedPosition{
			Name:  p.Name,
			Lon:   pos.Lon,
			Sign:  astromath.SignName(pos.Lon),
			Speed: pos.Speed,
			Retro: astromath.IsRetrograde(pos.Speed),
			House: astromath.HouseForLon(pos.Lon, chart.Cusps),
		}

		prevPos, err := ephemeris.CalcPlanet(prevProgressedJD, p.ID, flag)
		if err == nil {
			prevSign := astromath.SignName(prevPos.Lon)
			prevHouse := astromath.HouseForLon(prevPos.Lon, chart.Cusps)

			if prevSign != pp.Sign {
				pp.SignIngress = true
				pp.PrevSign = prevSign
			}
			if prevHouse != pp.House {
				pp.HouseIngress = true
				pp.PrevHouse = prevHouse
			}
		}

		positions = append(positions, pp)
	}

	y, _, _, _ := ephemeris.RevJul(jdTarget)
	return &ProgressionsResult{
		Year:         y,
		ProgressedJD: progressedJD,
		AgeYears:     ageYears,
		Positions:    positions,
	}, nil
}
