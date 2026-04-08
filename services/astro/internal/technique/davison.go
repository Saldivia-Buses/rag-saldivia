package technique

import (
	"time"

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
// Then builds a real natal chart at that moment and location.
func CalcDavison(pair *ChartPair, birthA, birthB time.Time) (*DavisonResult, error) {
	midJD := (pair.ChartA.JD + pair.ChartB.JD) / 2
	midLat := (pair.ChartA.Lat + pair.ChartB.Lat) / 2
	midLon := (pair.ChartA.Lon + pair.ChartB.Lon) / 2
	midAlt := (pair.ChartA.Alt + pair.ChartB.Alt) / 2

	// Extract the hour from the midpoint JD
	// midJD already encodes the time — we need to extract UT hour for BuildNatal
	// Since BuildNatal takes local hour + utcOffset, we pass UT hour with offset=0
	_, _, _, hourUT := revJulSimple(midJD)

	// Extract calendar date from midpoint JD
	year, month, day, _ := revJulSimple(midJD)

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

// revJulSimple is a simplified reverse Julian Day calculation.
// For the Davison chart we need calendar components from the midpoint JD.
// Uses the ephemeris package's RevJul internally.
func revJulSimple(jd float64) (int, int, int, float64) {
	// Meeus algorithm (Astronomical Algorithms, 1991)
	z := int(jd + 0.5)
	f := jd + 0.5 - float64(z)

	var a int
	if z < 2299161 {
		a = z
	} else {
		alpha := int((float64(z) - 1867216.25) / 36524.25)
		a = z + 1 + alpha - alpha/4
	}

	b := a + 1524
	c := int((float64(b) - 122.1) / 365.25)
	d := int(365.25 * float64(c))
	e := int(float64(b-d) / 30.6001)

	day := b - d - int(30.6001*float64(e))

	var month int
	if e < 14 {
		month = e - 1
	} else {
		month = e - 13
	}

	var year int
	if month > 2 {
		year = c - 4716
	} else {
		year = c - 4715
	}

	hour := f * 24.0

	return year, month, day, hour
}
