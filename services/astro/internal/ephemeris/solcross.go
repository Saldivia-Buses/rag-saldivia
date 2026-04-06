package ephemeris

import (
	"fmt"
	"math"
)

// SolcrossUT finds the Julian Day when the Sun's longitude crosses targetLon
// after tjdStart. Uses Newton iteration over CalcPlanet.
// Replaces pyswisseph's swe.solcross_ut which swephgo doesn't expose.
func SolcrossUT(targetLon, tjdStart float64, flags int) (float64, error) {
	const maxIter = 50
	const tolerance = 1e-7

	jd := tjdStart
	for iter := 0; iter < maxIter; iter++ {
		pos, err := CalcPlanet(jd, Sun, flags)
		if err != nil {
			return 0, err
		}
		diff := targetLon - pos.Lon
		for diff > 180 {
			diff -= 360
		}
		for diff < -180 {
			diff += 360
		}
		if math.Abs(diff) < tolerance {
			return jd, nil
		}
		speed := pos.Speed
		if speed == 0 {
			speed = 1.0
		}
		jd += diff / speed
	}
	return 0, fmt.Errorf("SolcrossUT: did not converge after %d iterations (target=%.4f)", maxIter, targetLon)
}
