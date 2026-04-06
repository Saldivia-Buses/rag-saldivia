package technique

import (
	"math"

	"github.com/Camionerou/rag-saldivia/services/astro/internal/astromath"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/ephemeris"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/natal"
)

// TransitActivation holds a slow-planet transit hitting a natal point.
type TransitActivation struct {
	Transit    string          `json:"transit"`
	Aspect     string          `json:"aspect"`
	Natal      string          `json:"natal"`
	Orb        float64         `json:"orb"`
	TrPos      string          `json:"tr_pos"`
	Retrograde bool            `json:"retrograde"`
	Passes     int             `json:"passes"`
	Month      int             `json:"month"`
	Nature     string          `json:"nature"`
	EpDetails  []EpisodeDetail `json:"ep_details"`
}

// EpisodeDetail summarizes a transit pass.
type EpisodeDetail struct {
	MonthStart int  `json:"month_start"`
	MonthEnd   int  `json:"month_end"`
	Retrograde bool `json:"retrograde"`
}

// transitSample holds a sampled position for a slow planet.
type transitSample struct {
	jd    float64
	lon   float64
	ra    float64
	speed float64
}

// slowPlanetIDs maps to the planets we sample for transits.
var slowPlanetIDs = []struct {
	Name string
	ID   int
}{
	{"Júpiter", ephemeris.Jupiter},
	{"Saturno", ephemeris.Saturn},
	{"Urano", ephemeris.Uranus},
	{"Neptuno", ephemeris.Neptune},
	{"Plutón", ephemeris.Pluto},
	{"Nodo Norte", ephemeris.TrueNode},
	{"Quirón", ephemeris.Chiron},
}

const (
	sampleDays     = 5   // sample every 5 days
	tripleGapDays  = 20  // >20 days apart = new episode
)

// CalcTransits samples slow planets every 5 days and finds aspects to natal points.
// Uses mundane aspects (equatorial RA) matching Python's transits_context().
func CalcTransits(chart *natal.Chart, year int) []TransitActivation {
	jdStart := ephemeris.JulDay(year, 1, 1, 0.0)
	jdEnd := ephemeris.JulDay(year+1, 1, 1, 0.0)
	flags := ephemeris.FlagSwieph | ephemeris.FlagSpeed
	orb := astromath.OrbDefaults.Transit

	// Collect natal points (RA for mundane aspects) including angles
	type natalPoint struct {
		name string
		ra   float64
	}
	var natPoints []natalPoint
	for name, pos := range chart.Planets {
		if pos.RA == 0 && pos.Dec == 0 && pos.Lon == 0 {
			continue
		}
		natPoints = append(natPoints, natalPoint{name, pos.RA})
	}
	// Add angles as transit targets (convert ecliptic lon → RA)
	ascRA, _ := astromath.EclToEq(chart.ASC, 0, chart.Epsilon)
	natPoints = append(natPoints, natalPoint{"ASC", ascRA})
	mcRA, _ := astromath.EclToEq(chart.MC, 0, chart.Epsilon)
	natPoints = append(natPoints, natalPoint{"MC", mcRA})
	if chart.Vertex != 0 {
		vtxRA, _ := astromath.EclToEq(chart.Vertex, 0, chart.Epsilon)
		natPoints = append(natPoints, natalPoint{"Vertex", vtxRA})
	}

	var results []TransitActivation

	for _, sp := range slowPlanetIDs {
		// Sample this planet every 5 days (single CalcPlanetFull call per sample)
		var samples []transitSample
		for jd := jdStart; jd < jdEnd; jd += sampleDays {
			pos, err := ephemeris.CalcPlanetFull(jd, sp.ID, flags)
			if err != nil {
				continue
			}
			samples = append(samples, transitSample{
				jd:    jd,
				lon:   pos.Lon,
				ra:    pos.RA,
				speed: pos.Speed,
			})
		}

		// Check each natal point for aspects
		for _, np := range natPoints {
			if sp.Name == np.name {
				continue // skip self
			}

			var hits []hit
			for _, s := range samples {
				asp := astromath.FindAspect(s.ra, np.ra, orb)
				if asp != nil {
					hits = append(hits, hit{
						jd:     s.jd,
						orb:    asp.Orb,
						aspect: asp.Name,
						retro:  s.speed < 0,
						trLon:  s.lon,
					})
				}
			}

			if len(hits) == 0 {
				continue
			}

			// Group hits into episodes (gap > 20 days = new episode)
			episodes := groupEpisodes(hits)

			// Find closest orb across all hits
			bestOrb := math.MaxFloat64
			bestAspect := ""
			bestLon := 0.0
			hasRetro := false
			bestMonth := 0
			for _, h := range hits {
				if h.orb < bestOrb {
					bestOrb = h.orb
					bestAspect = h.aspect
					bestLon = h.trLon
					_, m, _, _ := ephemeris.RevJul(h.jd)
					bestMonth = m
				}
				if h.retro {
					hasRetro = true
				}
			}

			// Build episode details
			var epDetails []EpisodeDetail
			for _, ep := range episodes {
				if len(ep) == 0 {
					continue
				}
				_, mStart, _, _ := ephemeris.RevJul(ep[0].jd)
				_, mEnd, _, _ := ephemeris.RevJul(ep[len(ep)-1].jd)
				epRetro := false
				for _, h := range ep {
					if h.retro {
						epRetro = true
						break
					}
				}
				epDetails = append(epDetails, EpisodeDetail{
					MonthStart: mStart, MonthEnd: mEnd, Retrograde: epRetro,
				})
			}

			results = append(results, TransitActivation{
				Transit:    sp.Name,
				Aspect:     bestAspect,
				Natal:      np.name,
				Orb:        bestOrb,
				TrPos:      astromath.PosToStr(bestLon),
				Retrograde: hasRetro,
				Passes:     len(episodes),
				Month:      bestMonth,
				Nature:     aspectNature(bestAspect),
				EpDetails:  epDetails,
			})
		}
	}

	return results
}

// groupEpisodes splits hits into episodes separated by >20 days.
// Assumes hits are ordered by JD ascending (guaranteed by the sampling loop).
func groupEpisodes(hits []hit) [][]hit {
	if len(hits) == 0 {
		return nil
	}
	var episodes [][]hit
	current := []hit{hits[0]}
	for _, h := range hits[1:] {
		if h.jd-current[len(current)-1].jd > tripleGapDays {
			episodes = append(episodes, current)
			current = []hit{h}
		} else {
			current = append(current, h)
		}
	}
	episodes = append(episodes, current)
	return episodes
}

// hit is used within CalcTransits for transit sampling.
type hit struct {
	jd     float64
	orb    float64
	aspect string
	retro  bool
	trLon  float64
}
