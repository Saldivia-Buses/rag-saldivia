package astromath

// FixedStar represents a major fixed star for astrological analysis.
type FixedStar struct {
	Name      string
	SweName   string
	Magnitude float64
	Nature    string
}

// MajorFixedStars — magnitude ≤ 1.5 stars with astrological significance.
// Ecliptic longitudes must be calculated via ephemeris.FixstarUT() for the date.
var MajorFixedStars = []FixedStar{
	{"Algol", "Algol", 2.1, "Saturn/Jupiter"},
	{"Aldebarán", "Aldebaran", 0.85, "Mars"},
	{"Rigel", "Rigel", 0.13, "Jupiter/Saturn"},
	{"Capella", "Capella", 0.08, "Mars/Mercury"},
	{"Sirio", "Sirius", -1.46, "Jupiter/Mars"},
	{"Cástor", "Castor", 1.58, "Mercury"},
	{"Pólux", "Pollux", 1.14, "Mars"},
	{"Régulo", "Regulus", 1.35, "Jupiter/Mars"},
	{"Spica", "Spica", 0.97, "Venus/Mars"},
	{"Arturo", "Arcturus", -0.05, "Jupiter/Mars"},
	{"Antares", "Antares", 1.09, "Mars/Jupiter"},
	{"Vega", "Vega", 0.03, "Venus/Mercury"},
	{"Altair", "Altair", 0.77, "Mars/Jupiter"},
	{"Fomalhaut", "Fomalhaut", 1.16, "Venus/Mercury"},
}

// FixedStarOrb — orb for fixed star conjunctions to natal points.
const FixedStarOrb = 1.5
