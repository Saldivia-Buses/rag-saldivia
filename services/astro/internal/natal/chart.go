package natal

import (
	"github.com/Camionerou/rag-saldivia/services/astro/internal/astromath"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/ephemeris"
)

// Chart holds a complete natal chart.
type Chart struct {
	JD         float64
	Lat        float64
	Lon        float64
	Alt        float64
	UTCOffset  int
	Planets    map[string]*ephemeris.PlanetPos
	Cusps      []float64 // [0] unused, [1]-[12] house cusps
	ASC        float64
	MC         float64
	ARMC       float64
	Vertex     float64
	Epsilon    float64
	Diurnal    bool
	Combustion map[string]string // planet → "combust"|"cazimi"|""
	Retrograde map[string]bool
}

// mainPlanets maps Spanish names to Swiss Ephemeris IDs for the 10 main planets.
var mainPlanets = map[string]int{
	"Sol": ephemeris.Sun, "Luna": ephemeris.Moon,
	"Mercurio": ephemeris.Mercury, "Venus": ephemeris.Venus,
	"Marte": ephemeris.Mars, "Júpiter": ephemeris.Jupiter,
	"Saturno": ephemeris.Saturn, "Urano": ephemeris.Uranus,
	"Neptuno": ephemeris.Neptune, "Plutón": ephemeris.Pluto,
}

// BuildNatal calculates a natal chart.
// Mirrors build_natal() from astro-v2/primary_directions.py.
// Uses Polich-Page Topocentric house system with topocentric planetary positions.
// Locks ephemeris.CalcMu for the entire calculation (SetTopo + all CalcPlanet must be atomic).
func BuildNatal(year, month, day int, localHour float64, lat, lon, alt float64, utcOffset int) (*Chart, error) {
	utHour := localHour - float64(utcOffset)
	jd := ephemeris.JulDay(year, month, day, utHour)

	ephemeris.CalcMu.Lock()
	defer ephemeris.CalcMu.Unlock()

	ephemeris.SetTopo(lon, lat, alt)

	epsilon, err := ephemeris.EclNut(jd)
	if err != nil {
		return nil, err
	}

	cusps, ascmc, err := ephemeris.CalcHouses(jd, lat, lon, ephemeris.HouseTopocentric)
	if err != nil {
		return nil, err
	}

	flag := ephemeris.FlagSwieph | ephemeris.FlagTopoctr | ephemeris.FlagSpeed
	planets := make(map[string]*ephemeris.PlanetPos)

	for name, pid := range mainPlanets {
		pos, err := ephemeris.CalcPlanetFullLocked(jd, pid, flag)
		if err != nil {
			return nil, err
		}
		planets[name] = pos
	}

	// Chiron (may fail without asteroid ephemeris files — skip gracefully)
	if pos, err := ephemeris.CalcPlanetFullLocked(jd, ephemeris.Chiron, flag); err == nil {
		planets["Quirón"] = pos
	}

	// True Node (North Node)
	if pos, err := ephemeris.CalcPlanetFullLocked(jd, ephemeris.TrueNode, flag); err == nil {
		planets["Nodo Norte"] = pos
		planets["Nodo Sur"] = &ephemeris.PlanetPos{
			Lon:   astromath.Normalize360(pos.Lon + 180),
			Lat:   -pos.Lat,
			Speed: pos.Speed,
			RA:    astromath.Normalize360(pos.RA + 180),
			Dec:   -pos.Dec,
		}
	}

	// Lilith (Mean Apogee) — geocentric, not topocentric
	if pos, err := ephemeris.CalcPlanetFullLocked(jd, ephemeris.MeanApog, ephemeris.FlagSwieph|ephemeris.FlagSpeed); err == nil {
		planets["Lilith"] = pos
	}

	asc := ascmc[0]
	sunLon := planets["Sol"].Lon
	moonLon := planets["Luna"].Lon

	// Diurnal = Sun above horizon (houses 7-12)
	sunHouse := astromath.HouseForLon(sunLon, cusps)
	diurnal := sunHouse >= 7

	// Part of Fortune + Part of Spirit
	planets["Fortuna"] = &ephemeris.PlanetPos{Lon: astromath.PartOfFortune(asc, moonLon, sunLon, diurnal)}
	planets["Espíritu"] = &ephemeris.PlanetPos{Lon: astromath.PartOfSpirit(asc, moonLon, sunLon, diurnal)}

	// Combustion + retrograde status
	combustion := make(map[string]string)
	retrograde := make(map[string]bool)
	skip := map[string]bool{"Sol": true, "Fortuna": true, "Espíritu": true, "Nodo Norte": true, "Nodo Sur": true}
	for name, pos := range planets {
		if skip[name] {
			continue
		}
		combustion[name] = astromath.CombustionStatus(pos.Lon, sunLon)
		retrograde[name] = astromath.IsRetrograde(pos.Speed)
	}

	return &Chart{
		JD:         jd,
		Lat:        lat,
		Lon:        lon,
		Alt:        alt,
		UTCOffset:  utcOffset,
		Planets:    planets,
		Cusps:      cusps,
		ASC:        asc,
		MC:         ascmc[1],
		ARMC:       ascmc[2],
		Vertex:     ascmc[3],
		Epsilon:    epsilon,
		Diurnal:    diurnal,
		Combustion: combustion,
		Retrograde: retrograde,
	}, nil
}

// HouseOf returns which house (1-12) a planet occupies.
func (c *Chart) HouseOf(planetName string) int {
	pos, ok := c.Planets[planetName]
	if !ok {
		return 0
	}
	return astromath.HouseForLon(pos.Lon, c.Cusps)
}

// BoundLord returns the Egyptian bound lord for a planet's longitude.
func (c *Chart) BoundLord(planetName string) string {
	pos, ok := c.Planets[planetName]
	if !ok {
		return ""
	}
	return astromath.BoundLord(pos.Lon)
}
