package astromath

// Signs — Spanish zodiac sign names.
var Signs = [12]string{
	"Aries", "Tauro", "Géminis", "Cáncer",
	"Leo", "Virgo", "Libra", "Escorpio",
	"Sagitario", "Capricornio", "Acuario", "Piscis",
}

// PlanetNames maps planet ID to Spanish name.
var PlanetNames = map[int]string{
	0: "Sol", 1: "Luna", 2: "Mercurio", 3: "Venus", 4: "Marte",
	5: "Júpiter", 6: "Saturno", 7: "Urano", 8: "Neptuno", 9: "Plutón",
	15: "Quirón", 11: "Nodo Norte", 12: "Lilith",
}

// PlanetIDByName maps Spanish name to Swiss Ephemeris planet ID.
var PlanetIDByName = map[string]int{
	"Sol": 0, "Luna": 1, "Mercurio": 2, "Venus": 3, "Marte": 4,
	"Júpiter": 5, "Saturno": 6, "Urano": 7, "Neptuno": 8, "Plutón": 9,
	"Quirón": 15, "Nodo Norte": 11, "Lilith": 12,
}

// NaibodRate is the solar mean motion in degrees per year (Naibod 1556).
const NaibodRate = 0.985626

// Standard aspect angles.
var AspectAngles = map[string]float64{
	"conjunction": 0, "sextile": 60, "square": 90,
	"trine": 120, "opposition": 180,
}

// AspectNames maps angle to Spanish name.
var AspectNames = map[float64]string{
	0: "conjunción", 60: "sextil", 90: "cuadratura",
	120: "trígono", 180: "oposición",
}

// Default orbs (degrees) — from astro-v2 config.py.
var OrbDefaults = struct {
	Transit     float64
	SolarArc    float64
	PrimaryDir  float64
	SolarReturn float64
	Progression float64
	Synastry    float64
	FastTransit float64
	Station     float64
	SAAntiscia  float64
}{
	Transit: 3.0, SolarArc: 1.5, PrimaryDir: 2.0, SolarReturn: 2.0,
	Progression: 1.5, Synastry: 4.0, FastTransit: 3.0, Station: 5.0,
	SAAntiscia: 1.0,
}

// SlowPlanets — only these get transit analysis (matching astro-v2 config).
var SlowPlanets = []string{"Júpiter", "Saturno", "Urano", "Neptuno", "Plutón", "Nodo Norte", "Quirón"}

// SignLord — traditional sign rulers for profections chronocrator.
var SignLord = map[int]string{
	0: "Marte", 1: "Venus", 2: "Mercurio", 3: "Luna",
	4: "Sol", 5: "Mercurio", 6: "Venus", 7: "Marte",
	8: "Júpiter", 9: "Saturno", 10: "Saturno", 11: "Júpiter",
}
