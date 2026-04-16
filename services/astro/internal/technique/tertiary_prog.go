package technique

import (
	"github.com/Camionerou/rag-saldivia/services/astro/internal/astromath"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/ephemeris"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/natal"
)

// LunarMonthDays is the sidereal lunar month used for tertiary progressions.
// 1 progressed day = 27.321661 real days.
const LunarMonthDays = 27.321661

// TertiaryResult holds tertiary progressed positions for a target year.
type TertiaryResult struct {
	Year         int                  `json:"year"`
	ProgressedJD float64              `json:"progressed_jd"`
	AgeDays      float64              `json:"age_days"`
	Positions    []ProgressedPosition `json:"positions"` // reuses SP type
}

// tertiaryPlanetOrder matches secondary progressions (Sun through Saturn).
var tertiaryPlanetOrder = []struct {
	Name string
	ID   int
}{
	{"Sol", ephemeris.Sun}, {"Luna", ephemeris.Moon},
	{"Mercurio", ephemeris.Mercury}, {"Venus", ephemeris.Venus},
	{"Marte", ephemeris.Mars}, {"Júpiter", ephemeris.Jupiter},
	{"Saturno", ephemeris.Saturn},
}

// CalcTertiaryProgressions calculates tertiary progressions for a target year.
// Rate: 1 progressed day = 1 lunar month (27.321661 days) of real life.
// prog_jd = natal_jd + (age_in_days / 27.321661)
func CalcTertiaryProgressions(chart *natal.Chart, targetYear int) (*TertiaryResult, error) {
	jdMid := ephemeris.JulDay(targetYear, 7, 1, 12.0)
	ageDays := jdMid - chart.JD // age in days (JD is already in days)

	progressedDays := ageDays / LunarMonthDays
	progressedJD := chart.JD + progressedDays

	// Also compute previous month's progressed JD for ingress detection
	prevAgeDays := ageDays - LunarMonthDays // one lunar month earlier
	prevProgressedDays := prevAgeDays / LunarMonthDays
	prevProgressedJD := chart.JD + prevProgressedDays

	ephemeris.CalcMu.Lock()
	defer ephemeris.CalcMu.Unlock()

	ephemeris.SetTopo(chart.Lon, chart.Lat, chart.Alt)

	flag := ephemeris.FlagSwieph | ephemeris.FlagTopoctr | ephemeris.FlagSpeed
	positions := make([]ProgressedPosition, 0, len(tertiaryPlanetOrder))

	for _, p := range tertiaryPlanetOrder {
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

		// Ingress detection vs previous month
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

	return &TertiaryResult{
		Year:         targetYear,
		ProgressedJD: progressedJD,
		AgeDays:      ageDays,
		Positions:    positions,
	}, nil
}
