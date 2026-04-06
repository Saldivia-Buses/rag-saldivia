package astromath

// HouseForLon returns the house number (1-12) for an ecliptic longitude.
// cusps must have at least 13 elements (Swiss Ephemeris convention: [0] unused, [1]-[12] houses).
func HouseForLon(lon float64, cusps []float64) int {
	if len(cusps) < 13 {
		return 1
	}
	lon = Normalize360(lon)
	for i := 1; i <= 12; i++ {
		next := i%12 + 1
		c1 := Normalize360(cusps[i])
		c2 := Normalize360(cusps[next])
		if c1 < c2 {
			if lon >= c1 && lon < c2 {
				return i
			}
		} else {
			if lon >= c1 || lon < c2 {
				return i
			}
		}
	}
	return 1
}
