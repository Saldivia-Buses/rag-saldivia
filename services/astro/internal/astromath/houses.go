package astromath

// HouseForLon returns the house number (1-12) for an ecliptic longitude.
func HouseForLon(lon float64, cusps []float64) int {
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
