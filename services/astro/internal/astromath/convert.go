package astromath

import "math"

// EclToEq converts ecliptic coordinates to equatorial (RA, Dec).
// eps is the obliquity of the ecliptic in degrees.
func EclToEq(lon, lat, eps float64) (ra, dec float64) {
	lonR := DegToRad(lon)
	latR := DegToRad(lat)
	epsR := DegToRad(eps)

	sinDec := math.Sin(latR)*math.Cos(epsR) + math.Cos(latR)*math.Sin(epsR)*math.Sin(lonR)
	dec = RadToDeg(math.Asin(sinDec))

	y := math.Sin(lonR)*math.Cos(epsR) - math.Tan(latR)*math.Sin(epsR)
	x := math.Cos(lonR)
	ra = RadToDeg(math.Atan2(y, x))
	ra = Normalize360(ra)
	return ra, dec
}

// PartOfFortune calculates the Part of Fortune.
// Diurnal: ASC + Moon - Sun. Nocturnal: ASC + Sun - Moon.
func PartOfFortune(asc, moonLon, sunLon float64, diurnal bool) float64 {
	if diurnal {
		return Normalize360(asc + moonLon - sunLon)
	}
	return Normalize360(asc + sunLon - moonLon)
}

// PartOfSpirit calculates the Part of Spirit (inverse of Fortune).
// Required for Zodiacal Releasing from Spirit.
func PartOfSpirit(asc, moonLon, sunLon float64, diurnal bool) float64 {
	if diurnal {
		return Normalize360(asc + sunLon - moonLon)
	}
	return Normalize360(asc + moonLon - sunLon)
}

// SouthNode returns the South Node longitude (180° from North Node).
func SouthNode(northNodeLon float64) float64 {
	return Normalize360(northNodeLon + 180)
}

// Antiscion returns the antiscion point (mirror across Cancer/Capricorn axis).
func Antiscion(lon float64) float64 {
	return Normalize360(180 - lon)
}

// ContraAntiscion returns the contra-antiscion (mirror across Aries/Libra axis).
func ContraAntiscion(lon float64) float64 {
	return Normalize360(360 - lon)
}

// Combustion thresholds (degrees from Sun).
const (
	OrbCombust = 8.0
	OrbCazimi  = 0.283 // 0°17'
)

// CombustionStatus returns "cazimi", "combust", or "" for a planet relative to the Sun.
func CombustionStatus(planetLon, sunLon float64) string {
	diff := AngDiff(planetLon, sunLon)
	if diff <= OrbCazimi {
		return "cazimi"
	}
	if diff <= OrbCombust {
		return "combust"
	}
	return ""
}
