package astromath

// GeoLocation holds coordinates for a city.
type GeoLocation struct {
	City    string
	Country string
	Lat     float64
	Lon     float64
	Alt     float64
	UTCOff  int
}

// ArgentineCities contains key Argentine cities for quick lookup.
// Avoids external API calls for the most common locations.
var ArgentineCities = map[string]GeoLocation{
	"buenos aires":   {"Buenos Aires", "Argentina", -34.6037, -58.3816, 25, -3},
	"rosario":        {"Rosario", "Argentina", -32.9468, -60.6393, 25, -3},
	"córdoba":        {"Córdoba", "Argentina", -31.4201, -64.1888, 390, -3},
	"cordoba":        {"Córdoba", "Argentina", -31.4201, -64.1888, 390, -3},
	"mendoza":        {"Mendoza", "Argentina", -32.8895, -68.8458, 750, -3},
	"tucumán":        {"Tucumán", "Argentina", -26.8083, -65.2176, 450, -3},
	"tucuman":        {"Tucumán", "Argentina", -26.8083, -65.2176, 450, -3},
	"mar del plata":  {"Mar del Plata", "Argentina", -38.0055, -57.5426, 21, -3},
	"salta":          {"Salta", "Argentina", -24.7821, -65.4232, 1187, -3},
	"santa fe":       {"Santa Fe", "Argentina", -31.6333, -60.7000, 18, -3},
	"la plata":       {"La Plata", "Argentina", -34.9215, -57.9545, 26, -3},
	"neuquén":        {"Neuquén", "Argentina", -38.9516, -68.0591, 271, -3},
	"neuquen":        {"Neuquén", "Argentina", -38.9516, -68.0591, 271, -3},
	"bahía blanca":   {"Bahía Blanca", "Argentina", -38.7196, -62.2724, 29, -3},
	"bahia blanca":   {"Bahía Blanca", "Argentina", -38.7196, -62.2724, 29, -3},
	"resistencia":    {"Resistencia", "Argentina", -27.4514, -58.9867, 52, -3},
	"posadas":        {"Posadas", "Argentina", -27.3621, -55.8961, 120, -3},
	"paraná":         {"Paraná", "Argentina", -31.7413, -60.5116, 78, -3},
	"parana":         {"Paraná", "Argentina", -31.7413, -60.5116, 78, -3},
	"san juan":       {"San Juan", "Argentina", -31.5375, -68.5364, 630, -3},
	"san luis":        {"San Luis", "Argentina", -33.3017, -66.3378, 715, -3},
}

// LookupCity returns coordinates for a known city (case-insensitive).
// Returns nil if not found — caller should fallback to external geocoding API.
func LookupCity(name string) *GeoLocation {
	// Normalize
	lower := ""
	for _, r := range name {
		if r >= 'A' && r <= 'Z' {
			lower += string(r + 32)
		} else {
			lower += string(r)
		}
	}
	if loc, ok := ArgentineCities[lower]; ok {
		return &loc
	}
	return nil
}
