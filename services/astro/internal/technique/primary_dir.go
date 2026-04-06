package technique

import (
	"math"
	"sort"

	"github.com/Camionerou/rag-saldivia/services/astro/internal/astromath"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/natal"
)

// PrimaryDirection holds a single primary direction activation.
type PrimaryDirection struct {
	Promissor    string  `json:"promissor"`
	Aspect       string  `json:"aspect"`
	Significator string  `json:"significator"`
	Arc          float64 `json:"arc"`
	AgeExact     float64 `json:"age_exact"`
	OrbDeg       float64 `json:"orb_deg"`
	Tipo         string  `json:"tipo"`     // "directa" or "conversa"
	Sistema      string  `json:"sistema"`  // "polich-page"
	Applying     bool    `json:"applying"`
}

// maxArcDeg is the maximum valid arc (~100 years).
const maxArcDeg = 100.0

// pdAspects maps Spanish aspect names to angles (major aspects).
var pdAspects = map[string]float64{
	"Conjunción": 0, "Sextil": 60, "Cuadratura": 90,
	"Trígono": 120, "Oposición": 180,
}

// clamp restricts a value to [-1, 1] for asin/acos safety.
func clamp(x float64) float64 {
	if x > 1 {
		return 1
	}
	if x < -1 {
		return -1
	}
	return x
}

// diurnalSemiArc returns the arc a planet travels above the horizon (0-180°).
// DSA = acos(-tan(dec) * tan(lat))
func diurnalSemiArc(decDeg, latDeg float64) float64 {
	val := clamp(-math.Tan(astromath.DegToRad(decDeg)) * math.Tan(astromath.DegToRad(latDeg)))
	return astromath.RadToDeg(math.Acos(val))
}

// meridianDistance returns the shortest arc from RA to the meridian (0-180°).
func meridianDistance(raDeg, ramcDeg float64) float64 {
	d := math.Abs(math.Mod(raDeg-ramcDeg, 360))
	if d > 180 {
		d = 360 - d
	}
	return d
}

// isAboveHorizon checks if a point is above the horizon.
func isAboveHorizon(raDeg, decDeg, ramcDeg, latDeg float64) bool {
	return meridianDistance(raDeg, ramcDeg) <= diurnalSemiArc(decDeg, latDeg)
}

// semiarcPole calculates the Polich-Page pole for a significator.
// pole = asin(sin(lat) * DM / SA)
func semiarcPole(raDeg, decDeg, ramcDeg, latDeg float64) float64 {
	md := meridianDistance(raDeg, ramcDeg)
	above := isAboveHorizon(raDeg, decDeg, ramcDeg, latDeg)
	dsa := diurnalSemiArc(decDeg, latDeg)

	var sa, dm float64
	if above {
		sa = dsa
		dm = md
	} else {
		sa = 180.0 - dsa
		dm = 180.0 - md
	}

	if math.Abs(sa) < 0.0001 {
		return 0.0
	}
	sinP := math.Sin(astromath.DegToRad(latDeg)) * dm / sa
	return astromath.RadToDeg(math.Asin(clamp(sinP)))
}

// obliqueAscension computes OA = RA - asin(tan(dec) * tan(pole)).
func obliqueAscension(raDeg, decDeg, poleDeg float64) float64 {
	val := clamp(math.Tan(astromath.DegToRad(decDeg)) * math.Tan(astromath.DegToRad(poleDeg)))
	ad := astromath.RadToDeg(math.Asin(val))
	return math.Mod(raDeg-ad+360, 360)
}

// mundanePosition computes the mundane longitude (ML) using proportional semi-arcs.
// MC=0°, DSC=90°, IC=180°, ASC=270°.
func mundanePosition(raDeg, decDeg, ramcDeg, latDeg float64) float64 {
	md := meridianDistance(raDeg, ramcDeg)
	dsa := diurnalSemiArc(decDeg, latDeg)
	nsa := 180.0 - dsa
	above := isAboveHorizon(raDeg, decDeg, ramcDeg, latDeg)
	raDiff := math.Mod(raDeg-ramcDeg+360, 360)
	western := raDiff <= 180

	var ml float64
	switch {
	case above && western: // Q1
		if dsa > 0.001 {
			ml = md / dsa * 90.0
		}
	case !above && western: // Q2
		if nsa > 0.001 {
			ml = 90.0 + (md-dsa)/nsa*90.0
		} else {
			ml = 90.0
		}
	case !above && !western: // Q3
		if nsa > 0.001 {
			ml = 180.0 + (180.0-md)/nsa*90.0
		} else {
			ml = 180.0
		}
	default: // Q4: above && !western
		if dsa > 0.001 {
			ml = 270.0 + (dsa-md)/dsa*90.0
		} else {
			ml = 270.0
		}
	}

	return math.Mod(ml, 360)
}

// mundaneAspectPoint computes Point Q on the equator at the mundane aspect angle.
// Returns (RA_Q, Dec_Q=0).
func mundaneAspectPoint(sigRA, sigDec, aspAngle float64, sign int, ramcDeg, latDeg float64) (float64, float64) {
	mlS := mundanePosition(sigRA, sigDec, ramcDeg, latDeg)
	mlQ := math.Mod(mlS+float64(sign)*aspAngle+360, 360)
	raQ := math.Mod(ramcDeg+mlQ, 360)
	return raQ, 0.0 // Dec = 0 (equator)
}

// pdArcDirect computes the direct arc: OA_P(pole_S) - OA_Q(pole_S).
func pdArcDirect(promRA, promDec, aspRA, aspDec, sigRA, sigDec, ramc, lat float64) (float64, bool) {
	pole := semiarcPole(sigRA, sigDec, ramc, lat)
	oaP := obliqueAscension(promRA, promDec, pole)
	oaQ := obliqueAscension(aspRA, aspDec, pole)
	arc := math.Mod(oaP-oaQ+360, 360)
	if arc > 0 && arc <= maxArcDeg {
		return arc, true
	}
	return 0, false
}

// pdArcConverse computes the converse arc: OA_Q(pole_S) - OA_P(pole_S).
func pdArcConverse(promRA, promDec, aspRA, aspDec, sigRA, sigDec, ramc, lat float64) (float64, bool) {
	pole := semiarcPole(sigRA, sigDec, ramc, lat)
	oaP := obliqueAscension(promRA, promDec, pole)
	oaQ := obliqueAscension(aspRA, aspDec, pole)
	arc := math.Mod(oaQ-oaP+360, 360)
	if arc > 0 && arc <= maxArcDeg {
		return arc, true
	}
	return 0, false
}

// FindDirections finds all Primary Directions active for a target age.
// Polich-Page Topocentric system — mundane aspects (Dec=0 point Q).
func FindDirections(chart *natal.Chart, targetAge, orbDeg float64) []PrimaryDirection {
	orbYears := orbDeg / astromath.NaibodRate
	ramc := chart.ARMC
	lat := chart.Lat

	// Collect promissors and significators (all planets + angles)
	type point struct {
		name string
		ra   float64
		dec  float64
	}

	var points []point
	for name, pos := range chart.Planets {
		if pos.RA == 0 && pos.Dec == 0 && pos.Lon == 0 {
			continue // skip calculated points without real positions
		}
		points = append(points, point{name, pos.RA, pos.Dec})
	}

	var results []PrimaryDirection

	for _, prom := range points {
		for _, sig := range points {
			for aspName, aspAngle := range pdAspects {
				for _, sign := range []int{1, -1} {
					// Skip redundant sign=-1 for conjunction and opposition
					if aspAngle == 0 && sign == -1 {
						continue
					}
					if aspAngle == 180 && sign == -1 {
						continue
					}

					aspRA, aspDec := mundaneAspectPoint(sig.ra, sig.dec, aspAngle, sign, ramc, lat)

					// Direct arc
					if arc, ok := pdArcDirect(prom.ra, prom.dec, aspRA, aspDec, sig.ra, sig.dec, ramc, lat); ok {
						ageExact := arc / astromath.NaibodRate
						diff := math.Abs(ageExact - targetAge)
						if diff <= orbYears {
							results = append(results, PrimaryDirection{
								Promissor:    prom.name,
								Aspect:       aspName,
								Significator: sig.name,
								Arc:          arc,
								AgeExact:     ageExact,
								OrbDeg:       diff * astromath.NaibodRate,
								Tipo:         "directa",
								Sistema:      "polich-page",
								Applying:     ageExact > targetAge,
							})
						}
					}

					// Converse arc
					if arc, ok := pdArcConverse(prom.ra, prom.dec, aspRA, aspDec, sig.ra, sig.dec, ramc, lat); ok {
						ageExact := arc / astromath.NaibodRate
						diff := math.Abs(ageExact - targetAge)
						if diff <= orbYears {
							results = append(results, PrimaryDirection{
								Promissor:    prom.name,
								Aspect:       aspName,
								Significator: sig.name,
								Arc:          arc,
								AgeExact:     ageExact,
								OrbDeg:       diff * astromath.NaibodRate,
								Tipo:         "conversa",
								Sistema:      "polich-page",
								Applying:     ageExact > targetAge,
							})
						}
					}
				}
			}
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].OrbDeg < results[j].OrbDeg
	})

	return results
}
