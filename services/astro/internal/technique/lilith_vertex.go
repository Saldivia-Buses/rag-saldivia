package technique

import (
	"github.com/Camionerou/rag-saldivia/services/astro/internal/astromath"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/ephemeris"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/natal"
)

// LilithVertexResult holds extended Lilith and Vertex transit analysis.
type LilithVertexResult struct {
	LilithNatal  LilithInfo    `json:"lilith_natal"`
	VertexNatal  VertexInfo    `json:"vertex_natal"`
	Transits     []LVTransit   `json:"transits,omitempty"`
}

type LilithInfo struct {
	Lon   float64 `json:"lon"`
	Sign  string  `json:"sign"`
	House int     `json:"house"`
}

type VertexInfo struct {
	Lon   float64 `json:"lon"`
	Sign  string  `json:"sign"`
	House int     `json:"house"`
}

type LVTransit struct {
	Planet  string  `json:"planet"`
	Target  string  `json:"target"` // "Lilith" or "Vertex"
	Aspect  string  `json:"aspect"`
	Orb     float64 `json:"orb"`
	Month   int     `json:"month"`
}

// CalcLilithVertex computes natal positions and yearly transits to Lilith and Vertex.
func CalcLilithVertex(chart *natal.Chart, year int) *LilithVertexResult {
	result := &LilithVertexResult{}

	if p, ok := chart.Planets["Lilith"]; ok {
		result.LilithNatal = LilithInfo{
			Lon: p.Lon, Sign: astromath.SignName(p.Lon),
			House: astromath.HouseForLon(p.Lon, chart.Cusps),
		}
	}
	result.VertexNatal = VertexInfo{
		Lon: chart.Vertex, Sign: astromath.SignName(chart.Vertex),
		House: astromath.HouseForLon(chart.Vertex, chart.Cusps),
	}

	// Transits to Lilith and Vertex
	targets := map[string]float64{"Lilith": result.LilithNatal.Lon, "Vertex": chart.Vertex}
	slowIDs := []struct{ name string; id int }{
		{"Júpiter", ephemeris.Jupiter}, {"Saturno", ephemeris.Saturn},
		{"Urano", ephemeris.Uranus}, {"Neptuno", ephemeris.Neptune}, {"Plutón", ephemeris.Pluto},
	}
	flags := ephemeris.FlagSwieph | ephemeris.FlagSpeed

	for month := 1; month <= 12; month++ {
		jd := ephemeris.JulDay(year, month, 15, 12.0)
		for _, sp := range slowIDs {
			pos, err := ephemeris.CalcPlanet(jd, sp.id, flags)
			if err != nil {
				continue
			}
			for tName, tLon := range targets {
				if tLon == 0 {
					continue
				}
				asp := astromath.FindAspect(pos.Lon, tLon, 3.0)
				if asp != nil {
					result.Transits = append(result.Transits, LVTransit{
						Planet: sp.name, Target: tName, Aspect: asp.Name,
						Orb: asp.Orb, Month: month,
					})
				}
			}
		}
	}

	return result
}
