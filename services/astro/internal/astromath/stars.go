package astromath

// FixedStar represents a major fixed star for astrological analysis.
type FixedStar struct {
	Name      string
	SweName   string
	Magnitude float64
	Nature    string
}

// MajorFixedStars — 60 fixed stars with astrological significance.
// Ecliptic longitudes must be calculated via ephemeris.FixstarUT() for the date.
// Sources: Brady's Book of Fixed Stars, Robson, Lilly. SweName must match sefstars.txt exactly.
var MajorFixedStars = []FixedStar{
	// ── Royal Stars + brightest (original 14) ──
	{"Algol", "Algol", 2.10, "Saturn/Jupiter"},
	{"Aldebarán", "Aldebaran", 0.85, "Mars"},
	{"Rigel", "Rigel", 0.13, "Jupiter/Saturn"},
	{"Capella", "Capella", 0.08, "Mars/Mercury"},
	{"Sirio", "Sirius", -1.46, "Jupiter/Mars"},
	{"Cástor", "Castor", 1.58, "Mars/Mercury"},
	{"Pólux", "Pollux", 1.14, "Mars"},
	{"Régulo", "Regulus", 1.35, "Jupiter/Mars"},
	{"Spica", "Spica", 0.97, "Venus/Mars"},
	{"Arturo", "Arcturus", -0.05, "Jupiter/Mars"},
	{"Antares", "Antares", 1.09, "Mars/Jupiter"},
	{"Vega", "Vega", 0.03, "Venus/Mercury"},
	{"Altair", "Altair", 0.77, "Mars/Jupiter"},
	{"Fomalhaut", "Fomalhaut", 1.16, "Venus/Mercury"},

	// ── Aries / Taurus ──
	{"Hamal", "Hamal", 2.00, "Saturn/Mars"},
	{"Sheratan", "Sheratan", 2.64, "Saturn/Mars"},
	{"Pléyades", "Alcyone", 2.87, "Moon/Mars"},
	{"Menkar", "Menkar", 2.53, "Saturn"},

	// ── Orion / Gemini / Canis ──
	{"Betelgeuse", "Betelgeuse", 0.42, "Mars/Mercury"},
	{"Bellatrix", "Bellatrix", 1.64, "Mars/Mercury"},
	{"Mintaka", "Mintaka", 2.23, "Saturn/Mercury"},
	{"Alhena", "Alhena", 1.93, "Mercury/Venus"},
	{"Canopus", "Canopus", -0.74, "Saturn/Jupiter"},
	{"Procyon", "Procyon", 0.34, "Mercury/Mars"},

	// ── Cancer / Leo ──
	{"Wasat", "Wasat", 3.53, "Saturn"},
	{"Algieba", "Algieba", 1.98, "Saturn/Venus"},
	{"Zosma", "Zosma", 2.56, "Saturn/Venus"},
	{"Denébola", "Denebola", 2.14, "Saturn/Venus"},

	// ── Virgo / Libra ──
	{"Vindemiatrix", "Vindemiatrix", 2.83, "Saturn/Mercury"},
	{"Zaniah", "Zaniah", 3.89, "Mercury/Venus"},
	{"Algorab", "Algorab", 2.94, "Mars/Saturn"},
	{"Zubenelgenubi", "Zubenelgenubi", 2.75, "Saturn/Mars"},
	{"Zubeneschamali", "Zubeneschamali", 2.61, "Jupiter/Mercury"},

	// ── Scorpio / Ophiuchus ──
	{"Acrab", "Acrab", 2.62, "Mars/Saturn"},
	{"Dschubba", "Dschubba", 2.32, "Mars/Saturn"},
	{"Sabik", "Sabik", 2.43, "Saturn/Venus"},
	{"Ras Alhague", "Rasalhague", 2.08, "Saturn/Venus"},

	// ── Sagittarius ──
	{"Facies", "Facies", 5.20, "Sun/Mars"},
	{"Nunki", "Nunki", 2.02, "Saturn/Jupiter"},
	{"Kaus Australis", "Kaus Australis", 1.85, "Jupiter/Mars"},
	{"Ascella", "Ascella", 2.60, "Jupiter/Mercury"},

	// ── Capricorn / Aquarius ──
	{"Deneb Algedi", "Deneb Algedi", 2.85, "Saturn/Jupiter"},
	{"Sadalmelik", "Sadalmelik", 2.96, "Saturn/Mercury"},
	{"Sadalsuud", "Sadalsuud", 2.91, "Saturn/Mercury"},
	{"Sadachbia", "Sadachbia", 3.84, "Saturn/Mercury"},

	// ── Pegasus / Pisces / Andromeda ──
	{"Markab", "Markab", 2.49, "Mars/Mercury"},
	{"Scheat", "Scheat", 2.42, "Mars/Mercury"},
	{"Enif", "Enif", 2.38, "Saturn/Jupiter"},
	{"Mirach", "Mirach", 2.07, "Venus"},
	{"Almach", "Almach", 2.10, "Venus/Mars"},
	{"Achernar", "Achernar", 0.46, "Jupiter"},
	{"Alphecca", "Alphecca", 2.23, "Venus/Mercury"},

	// ── Aquila / Cygnus / Lyra ──
	{"Tarazed", "Tarazed", 2.72, "Mars/Jupiter"},
	{"Deneb Adige", "Deneb Adige", 1.25, "Venus/Mercury"},

	// ── Southern Cross / Centaurus ──
	{"Rigil Kentaurus", "Rigil Kent", -0.27, "Venus/Jupiter"},
	{"Agena", "Agena", 0.61, "Venus/Jupiter"},
	{"Acrux", "Acrux", 1.33, "Jupiter"},
	{"Mimosa", "Mimosa", 1.25, "Venus/Jupiter"},
	{"Acamar", "Acamar", 2.88, "Jupiter"},
	{"Mira", "Mira", 6.47, "Saturn/Jupiter"},
}

// FixedStarOrb — orb for fixed star conjunctions to natal points.
const FixedStarOrb = 1.5
