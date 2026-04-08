package astromath

// Essential dignities tables — Ptolemaic system.
// Reference: al-Qabisi "Introducción a la Ciencia de los Juicios de las Estrellas";
// Vettius Valens "Anthologiarum"; Lilly "Christian Astrology" (Book I, cap. IV).
//
// Scoring hierarchy:
//   Domicilio:   5 pts
//   Exaltación:  4 pts
//   Triplicidad: 3 pts
//   Término:     2 pts
//   Faz/Decano:  1 pt

// DignityScore constants.
const (
	ScoreDomicile    = 5
	ScoreExaltation  = 4
	ScoreTriplicity  = 3
	ScoreTerm        = 2
	ScoreFace        = 1
	ScoreDetriment   = -5
	ScoreFall        = -4
)

// Domicile maps each classical planet to the sign indices it rules.
// Index 0=Aries … 11=Pisces.
var Domicile = map[string][]int{
	"Sol":      {4},     // Leo
	"Luna":     {3},     // Cáncer
	"Mercurio": {2, 5},  // Géminis, Virgo
	"Venus":    {1, 6},  // Tauro, Libra
	"Marte":    {0, 7},  // Aries, Escorpio
	"Júpiter":  {8, 11}, // Sagitario, Piscis
	"Saturno":  {9, 10}, // Capricornio, Acuario
}

// DomicileOf returns the domicile ruler for a sign index (0-11).
// Uses traditional rulership (no modern planets).
var DomicileOf = [12]string{
	"Marte",    // Aries
	"Venus",    // Tauro
	"Mercurio", // Géminis
	"Luna",     // Cáncer
	"Sol",      // Leo
	"Mercurio", // Virgo
	"Venus",    // Libra
	"Marte",    // Escorpio
	"Júpiter",  // Sagitario
	"Saturno",  // Capricornio
	"Saturno",  // Acuario
	"Júpiter",  // Piscis
}

// Detriment is the opposite of domicile.
var DetrimentOf = [12]string{
	"Venus",    // Aries (opp Libra=Venus)
	"Marte",    // Tauro (opp Escorpio=Marte)
	"Júpiter",  // Géminis (opp Sagitario=Júpiter)
	"Saturno",  // Cáncer (opp Capricornio=Saturno)
	"Saturno",  // Leo (opp Acuario=Saturno)
	"Júpiter",  // Virgo (opp Piscis=Júpiter)
	"Marte",    // Libra (opp Aries=Marte)
	"Venus",    // Escorpio (opp Tauro=Venus)
	"Mercurio", // Sagitario (opp Géminis=Mercurio)
	"Luna",     // Capricornio (opp Cáncer=Luna)
	"Sol",      // Acuario (opp Leo=Sol)
	"Mercurio", // Piscis (opp Virgo=Mercurio)
}

// ExaltationSign maps planet to (signIndex, exactDegree).
type ExaltationEntry struct {
	Sign   int
	Degree float64
}

var Exaltation = map[string]ExaltationEntry{
	"Sol":      {0, 19},  // Aries 19°
	"Luna":     {1, 3},   // Tauro 3°
	"Mercurio": {5, 15},  // Virgo 15°
	"Venus":    {11, 27}, // Piscis 27°
	"Marte":    {9, 28},  // Capricornio 28°
	"Júpiter":  {3, 15},  // Cáncer 15°
	"Saturno":  {6, 21},  // Libra 21°
}

// ExaltationOf returns the planet exalted in a given sign (empty if none).
var ExaltationOf = [12]string{
	"Sol",      // Aries
	"Luna",     // Tauro
	"",         // Géminis
	"Júpiter",  // Cáncer
	"",         // Leo
	"Mercurio", // Virgo
	"Saturno",  // Libra
	"",         // Escorpio
	"",         // Sagitario
	"Marte",    // Capricornio
	"",         // Acuario
	"Venus",    // Piscis
}

// FallOf returns the planet in fall in a given sign (opposite of exaltation).
var FallOf = [12]string{
	"Saturno",  // Aries (opp Libra)
	"",         // Tauro
	"",         // Géminis
	"Marte",    // Cáncer (opp Capricornio)
	"",         // Leo
	"Venus",    // Virgo (opp Piscis)
	"Sol",      // Libra (opp Aries)
	"Luna",     // Escorpio (opp Tauro)
	"",         // Sagitario
	"Júpiter",  // Capricornio (opp Cáncer)
	"",         // Acuario
	"Mercurio", // Piscis (opp Virgo)
}

// Triplicity maps element to (day ruler, night ruler) — Ptolemaic system.
// Fire: Aries(0), Leo(4), Sagitario(8)
// Earth: Tauro(1), Virgo(5), Capricornio(9)
// Air: Géminis(2), Libra(6), Acuario(10)
// Water: Cáncer(3), Escorpio(7), Piscis(11)
type TriplicityEntry struct {
	Day   string
	Night string
}

var signElement = [12]string{
	"fire", "earth", "air", "water",
	"fire", "earth", "air", "water",
	"fire", "earth", "air", "water",
}

var Triplicity = map[string]TriplicityEntry{
	"fire":  {"Sol", "Júpiter"},
	"earth": {"Venus", "Luna"},
	"air":   {"Saturno", "Mercurio"},
	"water": {"Venus", "Marte"},
}

// TriplicityLord returns the triplicity lord for a sign index given sect.
func TriplicityLord(signIdx int, diurnal bool) string {
	elem := signElement[signIdx%12]
	trip, ok := Triplicity[elem]
	if !ok {
		return ""
	}
	if diurnal {
		return trip.Day
	}
	return trip.Night
}

// PtolemaicTerms maps sign name to a list of (upperDegree, lord).
// The term applies from the previous upper bound to this one (exclusive).
// Each sign has 5 terms totaling 30°.
type TermEntry struct {
	UpperDeg float64
	Lord     string
}

var PtolemaicTerms = map[int][]TermEntry{
	0:  {{6, "Júpiter"}, {12, "Venus"}, {20, "Mercurio"}, {25, "Marte"}, {30, "Saturno"}},     // Aries
	1:  {{8, "Venus"}, {14, "Mercurio"}, {22, "Júpiter"}, {27, "Saturno"}, {30, "Marte"}},      // Tauro
	2:  {{6, "Mercurio"}, {12, "Júpiter"}, {17, "Venus"}, {24, "Marte"}, {30, "Saturno"}},      // Géminis
	3:  {{7, "Marte"}, {13, "Venus"}, {19, "Mercurio"}, {26, "Júpiter"}, {30, "Saturno"}},      // Cáncer
	4:  {{6, "Júpiter"}, {11, "Venus"}, {18, "Saturno"}, {24, "Mercurio"}, {30, "Marte"}},      // Leo
	5:  {{7, "Mercurio"}, {17, "Venus"}, {21, "Júpiter"}, {28, "Marte"}, {30, "Saturno"}},      // Virgo
	6:  {{6, "Saturno"}, {14, "Mercurio"}, {21, "Júpiter"}, {28, "Venus"}, {30, "Marte"}},      // Libra
	7:  {{7, "Marte"}, {11, "Venus"}, {19, "Mercurio"}, {24, "Júpiter"}, {30, "Saturno"}},      // Escorpio
	8:  {{12, "Júpiter"}, {17, "Venus"}, {21, "Mercurio"}, {26, "Saturno"}, {30, "Marte"}},     // Sagitario
	9:  {{7, "Mercurio"}, {14, "Júpiter"}, {22, "Venus"}, {26, "Saturno"}, {30, "Marte"}},      // Capricornio
	10: {{7, "Mercurio"}, {13, "Venus"}, {20, "Júpiter"}, {25, "Marte"}, {30, "Saturno"}},      // Acuario
	11: {{12, "Venus"}, {16, "Júpiter"}, {19, "Mercurio"}, {28, "Marte"}, {30, "Saturno"}},     // Piscis
}

// TermLord returns the term lord for a given ecliptic longitude.
func TermLord(lon float64) string {
	lon = Normalize360(lon)
	signIdx := int(lon/30) % 12
	deg := lon - float64(signIdx*30)
	terms := PtolemaicTerms[signIdx]
	for _, t := range terms {
		if deg < t.UpperDeg {
			return t.Lord
		}
	}
	if len(terms) > 0 {
		return terms[len(terms)-1].Lord
	}
	return ""
}

// FaceLord returns the face (decan) lord for a given ecliptic longitude.
// Chaldean order starting from Mars at 0° Aries.
var chaldeanOrder = [7]string{"Marte", "Sol", "Venus", "Mercurio", "Luna", "Saturno", "Júpiter"}

func FaceLord(lon float64) string {
	lon = Normalize360(lon)
	decanIdx := int(lon/10) % 36
	return chaldeanOrder[decanIdx%7]
}

// DignityScoreAt returns the highest essential dignity score a planet has at a longitude.
// Returns (score, dignityName). If the planet has no dignity, returns (0, "").
// Also checks debilities: detriment (-5) and fall (-4).
func DignityScoreAt(planet string, lon float64, diurnal bool) (int, string) {
	lon = Normalize360(lon)
	signIdx := int(lon/30) % 12

	// Domicile (5)
	if signs, ok := Domicile[planet]; ok {
		for _, s := range signs {
			if s == signIdx {
				return ScoreDomicile, "domicilio"
			}
		}
	}

	// Exaltation (4)
	if ex, ok := Exaltation[planet]; ok {
		if ex.Sign == signIdx {
			return ScoreExaltation, "exaltación"
		}
	}

	// Triplicity (3)
	if TriplicityLord(signIdx, diurnal) == planet {
		return ScoreTriplicity, "triplicidad"
	}

	// Term (2)
	if TermLord(lon) == planet {
		return ScoreTerm, "término"
	}

	// Face/Decan (1)
	if FaceLord(lon) == planet {
		return ScoreFace, "faz"
	}

	// Check debilities
	if DetrimentOf[signIdx] == planet {
		return ScoreDetriment, "detrimento"
	}
	if FallOf[signIdx] == planet {
		return ScoreFall, "caída"
	}

	return 0, ""
}

// ClassicalPlanets are the 7 traditional planets used for dignity calculations.
var ClassicalPlanets = [7]string{"Sol", "Luna", "Mercurio", "Venus", "Marte", "Júpiter", "Saturno"}
