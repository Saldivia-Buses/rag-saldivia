package technique

import (
	"time"

	"github.com/Camionerou/rag-saldivia/services/astro/internal/ephemeris"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/natal"
)

// DavisonResult holds a Davison Relationship Chart.
// Unlike composite (midpoint of positions), Davison is a REAL chart
// calculated for the midpoint date and midpoint location.
type DavisonResult struct {
	NameA    string       `json:"name_a"`
	NameB    string       `json:"name_b"`
	MidJD    float64      `json:"mid_jd"`
	MidLat   float64      `json:"mid_lat"`
	MidLon   float64      `json:"mid_lon"`
	MidAlt   float64      `json:"mid_alt"`
	Chart    *natal.Chart `json:"chart"`
}

// CalcDavison calculates a Davison Relationship Chart.
// Midpoint in TIME: (JD_A + JD_B) / 2
// Midpoint in SPACE: (lat_A + lat_B) / 2, (lon_A + lon_B) / 2
// Then builds a real natal chart at that moment and location via Swiss Ephemeris.
func CalcDavison(pair *ChartPair, birthA, birthB time.Time) (*DavisonResult, error) {
	midJD := (pair.ChartA.JD + pair.ChartB.JD) / 2
	midLat := (pair.ChartA.Lat + pair.ChartB.Lat) / 2
	midLon := (pair.ChartA.Lon + pair.ChartB.Lon) / 2
	midAlt := (pair.ChartA.Alt + pair.ChartB.Alt) / 2

	// Use Swiss Ephemeris RevJul for astronomically correct JD → calendar conversion
	year, month, day, hourUT := ephemeris.RevJul(midJD)

	// Build a real chart at the midpoint date/location
	// hourUT is already in UT, so utcOffset = 0
	chart, err := natal.BuildNatal(year, month, day, hourUT, midLat, midLon, midAlt, 0)
	if err != nil {
		return nil, err
	}

	return &DavisonResult{
		NameA:  pair.NameA,
		NameB:  pair.NameB,
		MidJD:  midJD,
		MidLat: midLat,
		MidLon: midLon,
		MidAlt: midAlt,
		Chart:  chart,
	}, nil
}
