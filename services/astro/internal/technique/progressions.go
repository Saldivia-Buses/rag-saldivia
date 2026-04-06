package technique

import (
	"github.com/Camionerou/rag-saldivia/services/astro/internal/astromath"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/ephemeris"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/natal"
)

// ProgressedPosition holds a progressed planet's position and ingress info.
type ProgressedPosition struct {
	Name      string  `json:"name"`
	Lon       float64 `json:"lon"`
	Sign      string  `json:"sign"`
	Speed     float64 `json:"speed"`
	Retro     bool    `json:"retrograde"`
	House     int     `json:"house"`
	Ingress   string  `json:"ingress,omitempty"`   // "sign" or "house" if changed
	PrevSign  string  `json:"prev_sign,omitempty"`
	PrevHouse int     `json:"prev_house,omitempty"`
}

// ProgressionsResult holds all progressed positions for a target year.
type ProgressionsResult struct {
	Year        int                   `json:"year"`
	ProgressedJD float64             `json:"progressed_jd"`
	AgeDays     float64              `json:"age_days"`
	Positions   []ProgressedPosition `json:"positions"`
}

// progressedPlanets are planets calculated in secondary progressions.
var progressedPlanets = map[string]int{
	"Sol": ephemeris.Sun, "Luna": ephemeris.Moon,
	"Mercurio": ephemeris.Mercury, "Venus": ephemeris.Venus,
	"Marte": ephemeris.Mars, "Júpiter": ephemeris.Jupiter,
	"Saturno": ephemeris.Saturn,
}

// CalcProgressions calculates Secondary Progressions for a target year.
// Day-for-year: 1 day of ephemeris = 1 year of life.
// progressed_jd = natal_jd + age_in_years (each year = 1 day)
func CalcProgressions(chart *natal.Chart, targetYear int) (*ProgressionsResult, error) {
	// Age in years at mid-year
	jdMid := ephemeris.JulDay(targetYear, 7, 1, 12.0)
	ageYears := (jdMid - chart.JD) / 365.25

	// Progressed JD: natal JD + age in days (1 day = 1 year)
	progressedJD := chart.JD + ageYears

	// Also calculate previous year for ingress detection
	prevProgressedJD := chart.JD + (ageYears - 1.0)

	flag := ephemeris.FlagSwieph | ephemeris.FlagSpeed
	var positions []ProgressedPosition

	for name, pid := range progressedPlanets {
		pos, err := ephemeris.CalcPlanet(progressedJD, pid, flag)
		if err != nil {
			continue
		}

		pp := ProgressedPosition{
			Name:  name,
			Lon:   pos.Lon,
			Sign:  astromath.SignName(pos.Lon),
			Speed: pos.Speed,
			Retro: astromath.IsRetrograde(pos.Speed),
			House: astromath.HouseForLon(pos.Lon, chart.Cusps),
		}

		// Ingress detection: compare with previous year's progressed position
		prevPos, err := ephemeris.CalcPlanet(prevProgressedJD, pid, flag)
		if err == nil {
			prevSign := astromath.SignName(prevPos.Lon)
			prevHouse := astromath.HouseForLon(prevPos.Lon, chart.Cusps)

			if prevSign != pp.Sign {
				pp.Ingress = "sign"
				pp.PrevSign = prevSign
			} else if prevHouse != pp.House {
				pp.Ingress = "house"
				pp.PrevHouse = prevHouse
			}
		}

		positions = append(positions, pp)
	}

	return &ProgressionsResult{
		Year:         targetYear,
		ProgressedJD: progressedJD,
		AgeDays:      ageYears,
		Positions:    positions,
	}, nil
}
