package astromath

// MelothesiaResult holds zodiac-to-body-part mapping for health analysis.
type MelothesiaResult struct {
	Mappings []BodyZoneMapping `json:"mappings"`
}

// BodyZoneMapping connects a sign to its body zone and any natal planets there.
type BodyZoneMapping struct {
	Sign       string   `json:"sign"`
	BodyZone   string   `json:"body_zone"`
	Planets    []string `json:"planets,omitempty"` // natal planets in this sign
	HasMalefic bool     `json:"has_malefic"`       // Mars or Saturn present
	Afflicted  bool     `json:"afflicted"`          // malefic + hard aspects to this zone
}

// MelothesiaZones maps each sign to its governed body zone (Manilius tradition).
var MelothesiaZones = [12]string{
	"cabeza, rostro, cerebro",                        // Aries
	"cuello, garganta, tiroides",                     // Tauro
	"brazos, manos, pulmones, sistema nervioso",      // Géminis
	"pecho, estómago, senos",                         // Cáncer
	"corazón, espalda, columna vertebral",            // Leo
	"intestinos, sistema digestivo, páncreas",        // Virgo
	"riñones, zona lumbar, piel",                     // Libra
	"órganos reproductivos, colon, sistema excretor", // Escorpio
	"caderas, muslos, hígado",                        // Sagitario
	"rodillas, huesos, articulaciones, dientes",      // Capricornio
	"tobillos, pantorrillas, sistema circulatorio",   // Acuario
	"pies, sistema linfático, glándula pineal",       // Piscis
}

// maleficPlanets are planets that can indicate health vulnerability.
var maleficPlanets = map[string]bool{"Marte": true, "Saturno": true, "Plutón": true}

// CalcMelothesia maps natal planet positions to body zones.
func CalcMelothesia(planetLons map[string]float64) *MelothesiaResult {
	// Group planets by sign
	signPlanets := make(map[int][]string)
	for name, lon := range planetLons {
		signIdx := SignIndex(lon)
		signPlanets[signIdx] = append(signPlanets[signIdx], name)
	}

	mappings := make([]BodyZoneMapping, 12)
	for i := 0; i < 12; i++ {
		planets := signPlanets[i]
		hasMalefic := false
		for _, p := range planets {
			if maleficPlanets[p] {
				hasMalefic = true
				break
			}
		}
		mappings[i] = BodyZoneMapping{
			Sign:       Signs[i],
			BodyZone:   MelothesiaZones[i],
			Planets:    planets,
			HasMalefic: hasMalefic,
			Afflicted:  hasMalefic && len(planets) > 1, // simplified: malefic + congestion
		}
	}

	return &MelothesiaResult{Mappings: mappings}
}
