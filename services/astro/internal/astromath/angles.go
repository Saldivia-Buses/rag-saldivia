package astromath

import (
	"fmt"
	"math"
)

// AngDiff returns the shortest angular distance between two longitudes (0-180°).
func AngDiff(a, b float64) float64 {
	d := math.Abs(a - b)
	if d > 180 {
		d = 360 - d
	}
	return d
}

// SignedDiff returns (b - a) normalized to [-180, +180].
func SignedDiff(a, b float64) float64 {
	d := math.Mod(b-a, 360)
	if d > 180 {
		d -= 360
	}
	if d < -180 {
		d += 360
	}
	return d
}

// Normalize360 normalizes an angle to [0, 360).
func Normalize360(deg float64) float64 {
	deg = math.Mod(deg, 360)
	if deg < 0 {
		deg += 360
	}
	return deg
}

// SignIndex returns the zodiac sign index (0-11) for an ecliptic longitude.
func SignIndex(lon float64) int {
	return int(Normalize360(lon) / 30.0)
}

// SignName returns the Spanish zodiac sign name for a longitude.
func SignName(lon float64) string {
	return Signs[SignIndex(lon)]
}

// PosToStr formats an ecliptic longitude as "D°MM' SignName".
func PosToStr(lon float64) string {
	lon = Normalize360(lon)
	sign := int(lon / 30.0)
	inSign := lon - float64(sign*30)
	deg := int(inSign)
	min := int((inSign - float64(deg)) * 60)
	return fmt.Sprintf("%d°%02d' %s", deg, min, Signs[sign])
}

// DegToRad converts degrees to radians.
func DegToRad(d float64) float64 { return d * math.Pi / 180 }

// RadToDeg converts radians to degrees.
func RadToDeg(r float64) float64 { return r * 180 / math.Pi }

// IsRetrograde returns true if the planet's speed is negative.
func IsRetrograde(speed float64) bool { return speed < 0 }
