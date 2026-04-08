package technique

import (
	"math"

	"github.com/Camionerou/rag-saldivia/services/astro/internal/astromath"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/ephemeris"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/natal"
)

// ACGResult holds astrocartography planet lines.
type ACGResult struct {
	Lines []ACGLine `json:"lines"`
}

// ACGLine represents one planet line on the map.
type ACGLine struct {
	Planet   string     `json:"planet"`
	LineType string     `json:"line_type"` // "MC", "IC", "ASC", "DSC"
	Points   []ACGPoint `json:"points"`    // lat/lon pairs along the line
}

// ACGPoint is a single point on a planet line.
type ACGPoint struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

// acgPlanets to compute lines for.
var acgPlanets = []struct {
	Name string
	ID   int
}{
	{"Sol", ephemeris.Sun}, {"Luna", ephemeris.Moon},
	{"Mercurio", ephemeris.Mercury}, {"Venus", ephemeris.Venus},
	{"Marte", ephemeris.Mars}, {"Júpiter", ephemeris.Jupiter},
	{"Saturno", ephemeris.Saturn}, {"Urano", ephemeris.Uranus},
	{"Neptuno", ephemeris.Neptune}, {"Plutón", ephemeris.Pluto},
}

// CalcAstrocartography computes MC/IC/ASC/DSC lines for all planets.
// MC line: longitudes where the planet is on the MC (upper meridian)
// IC line: longitudes where the planet is on the IC (lower meridian)
// ASC/DSC lines: curved lines where the planet rises/sets
// Resolution: gridRes degrees of longitude (default 5°).
func CalcAstrocartography(chart *natal.Chart, gridRes float64) *ACGResult {
	if gridRes <= 0 {
		gridRes = 5.0
	}
	result := &ACGResult{}

	for _, p := range acgPlanets {
		pos, ok := chart.Planets[p.Name]
		if !ok {
			continue
		}

		// MC line: where RAMC = planet's RA
		// At any longitude, the MC corresponds to a specific RA.
		// MC line longitude = planet RA converted to geographic longitude
		// MC_lon = -(RAMC - planet_RA) → this gives one longitude
		mcLon := astromath.Normalize360(-(chart.ARMC - pos.RA))
		if mcLon > 180 {
			mcLon -= 360
		}

		var mcPoints, icPoints []ACGPoint
		for lat := -70.0; lat <= 70.0; lat += gridRes {
			mcPoints = append(mcPoints, ACGPoint{Lat: lat, Lon: mcLon})
			icLon := mcLon + 180
			if icLon > 180 {
				icLon -= 360
			}
			icPoints = append(icPoints, ACGPoint{Lat: lat, Lon: icLon})
		}

		result.Lines = append(result.Lines,
			ACGLine{Planet: p.Name, LineType: "MC", Points: mcPoints},
			ACGLine{Planet: p.Name, LineType: "IC", Points: icPoints},
		)

		// ASC/DSC lines: planet on the eastern/western horizon
		// These are curved because rising/setting depends on latitude.
		// For each latitude, compute the longitude where the planet rises/sets.
		var ascPoints, dscPoints []ACGPoint
		for lat := -65.0; lat <= 65.0; lat += gridRes {
			// ASC longitude: where the planet's ecliptic longitude is on the eastern horizon
			// Simplified: use the oblique ascension formula
			ascLon, ok := calcRiseLon(pos.Dec, lat, pos.RA, true)
			if ok {
				ascPoints = append(ascPoints, ACGPoint{Lat: lat, Lon: ascLon})
			}
			dscLon, ok := calcRiseLon(pos.Dec, lat, pos.RA, false)
			if ok {
				dscPoints = append(dscPoints, ACGPoint{Lat: lat, Lon: dscLon})
			}
		}

		if len(ascPoints) > 0 {
			result.Lines = append(result.Lines, ACGLine{Planet: p.Name, LineType: "ASC", Points: ascPoints})
		}
		if len(dscPoints) > 0 {
			result.Lines = append(result.Lines, ACGLine{Planet: p.Name, LineType: "DSC", Points: dscPoints})
		}
	}

	return result
}

// calcRiseLon computes the geographic longitude where a planet with given
// declination rises (isRise=true) or sets (isRise=false) at a given latitude.
// Returns (longitude, ok). ok=false if the planet is circumpolar at that latitude.
func calcRiseLon(dec, lat, ra float64, isRise bool) (float64, bool) {
	decR := astromath.DegToRad(dec)
	latR := astromath.DegToRad(lat)

	// Hour angle at rise/set: cos(H) = -tan(dec)*tan(lat)
	cosH := -math.Tan(decR) * math.Tan(latR)
	if cosH < -1 || cosH > 1 {
		return 0, false // circumpolar — never rises/sets at this latitude
	}

	H := astromath.RadToDeg(math.Acos(cosH))

	var lon float64
	if isRise {
		// Rise: Local Sidereal Time = RA - H
		// Geographic longitude offset from Greenwich
		lon = -(ra - H)
	} else {
		lon = -(ra + H)
	}

	// Normalize to -180..+180
	lon = math.Mod(lon+540, 360) - 180

	return lon, true
}
