package technique

import (
	"time"

	"github.com/Camionerou/rag-saldivia/services/astro/internal/astromath"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/ephemeris"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/natal"
)

// HoraryResult holds a horary chart analysis.
type HoraryResult struct {
	Chart        *natal.Chart         `json:"chart"`
	QuestionTime time.Time            `json:"question_time"`
	VOCMoon      bool                 `json:"voc_moon"`        // void-of-course
	Radicality   string               `json:"radicality"`      // "radical" or "not radical"
	RadicalReason string              `json:"radical_reason"`
	ASCRuler     string               `json:"asc_ruler"`       // querent's significator
	ASCRulerHouse int                 `json:"asc_ruler_house"`
	MoonSign     string               `json:"moon_sign"`
	MoonHouse    int                  `json:"moon_house"`
	MoonLastAsp  *HoraryAspect        `json:"moon_last_aspect,omitempty"`
	MoonNextAsp  *HoraryAspect        `json:"moon_next_aspect,omitempty"`
	Considerations []string           `json:"considerations"`  // "before judgment" warnings
}

// HoraryAspect records Moon's last/next aspect.
type HoraryAspect struct {
	Planet string  `json:"planet"`
	Aspect string  `json:"aspect"`
	Orb    float64 `json:"orb"`
}

// CastHorary builds a horary chart for the moment a question is asked.
// Unlike natal techniques, there is no stored contact — the chart is ephemeral.
func CastHorary(questionTime time.Time, lat, lon, alt float64, utcOffset int) (*HoraryResult, error) {
	hour := float64(questionTime.Hour()) + float64(questionTime.Minute())/60.0

	chart, err := natal.BuildNatal(
		questionTime.Year(), int(questionTime.Month()), questionTime.Day(),
		hour, lat, lon, alt, utcOffset,
	)
	if err != nil {
		return nil, err
	}

	result := &HoraryResult{
		Chart:        chart,
		QuestionTime: questionTime,
	}

	// ASC ruler = querent's significator
	ascSign := astromath.SignIndex(chart.ASC)
	result.ASCRuler = astromath.DomicileOf[ascSign]
	if p, ok := chart.Planets[result.ASCRuler]; ok {
		result.ASCRulerHouse = astromath.HouseForLon(p.Lon, chart.Cusps)
	}

	// Moon analysis
	if moonPos, ok := chart.Planets["Luna"]; ok {
		result.MoonSign = astromath.SignName(moonPos.Lon)
		result.MoonHouse = astromath.HouseForLon(moonPos.Lon, chart.Cusps)

		// Void-of-course: Moon makes no major aspects before leaving its sign
		result.VOCMoon = isVOC(moonPos, chart)

		// Moon's last and next aspects
		result.MoonLastAsp, result.MoonNextAsp = moonAspects(moonPos, chart)
	}

	// Radicality check: ASC ruler matches the question's nature
	// Simplified: chart is radical if ASC degree is not in early (< 3°) or late (> 27°) degrees
	ascDeg := chart.ASC - float64(ascSign)*30
	if ascDeg < 3 {
		result.Radicality = "not radical"
		result.RadicalReason = "ASC en grados tempranos (<3°) — pregunta prematura"
		result.Considerations = append(result.Considerations, "ASC < 3° — la situación no está madura")
	} else if ascDeg > 27 {
		result.Radicality = "not radical"
		result.RadicalReason = "ASC en grados tardíos (>27°) — asunto ya decidido"
		result.Considerations = append(result.Considerations, "ASC > 27° — el asunto ya se resolvió o la pregunta es tardía")
	} else {
		result.Radicality = "radical"
		result.RadicalReason = "Carta radical — válida para juicio"
	}

	// Saturn in 7th = warning (astrologer may be hampered)
	if satPos, ok := chart.Planets["Saturno"]; ok {
		satHouse := astromath.HouseForLon(satPos.Lon, chart.Cusps)
		if satHouse == 7 {
			result.Considerations = append(result.Considerations, "Saturno en Casa 7 — precaución en el juicio")
		}
	}

	// VOC Moon consideration
	if result.VOCMoon {
		result.Considerations = append(result.Considerations, "Luna vacía de curso — nada llegará a concretarse")
	}

	return result, nil
}

// isVOC checks if the Moon is void-of-course (no major aspects before sign change).
func isVOC(moonPos *ephemeris.PlanetPos, chart *natal.Chart) bool {
	moonSign := astromath.SignIndex(moonPos.Lon)
	degsRemaining := 30.0 - (moonPos.Lon - float64(moonSign)*30)

	// How many degrees until sign change
	// Check if Moon makes any major aspect within those degrees
	for name, pos := range chart.Planets {
		if name == "Luna" {
			continue
		}
		// Check if Moon will aspect this planet before sign change
		for _, aspAngle := range []float64{0, 60, 90, 120, 180} {
			targetLon := astromath.Normalize360(pos.Lon + aspAngle)
			diff := astromath.SignedDiff(moonPos.Lon, targetLon)
			// Moon must be approaching (diff > 0) and within remaining degrees
			if diff > 0 && diff < degsRemaining {
				return false // Moon will make an aspect — not VOC
			}
			// Also check the other direction for the same aspect
			targetLon2 := astromath.Normalize360(pos.Lon - aspAngle)
			diff2 := astromath.SignedDiff(moonPos.Lon, targetLon2)
			if diff2 > 0 && diff2 < degsRemaining {
				return false
			}
		}
	}
	return true
}

// moonAspects finds Moon's most recent past aspect and next future aspect.
func moonAspects(moonPos *ephemeris.PlanetPos, chart *natal.Chart) (*HoraryAspect, *HoraryAspect) {
	var lastAsp, nextAsp *HoraryAspect
	bestLastDiff := 360.0
	bestNextDiff := 360.0

	for name, pos := range chart.Planets {
		if name == "Luna" {
			continue
		}
		asp := astromath.FindAspectWithMotion(moonPos.Lon, pos.Lon, moonPos.Speed, pos.Speed, 10.0)
		if asp == nil {
			continue
		}
		diff := astromath.AngDiff(moonPos.Lon, pos.Lon)

		if asp.Applying && diff < bestNextDiff {
			bestNextDiff = diff
			nextAsp = &HoraryAspect{Planet: name, Aspect: asp.Name, Orb: asp.Orb}
		}
		if !asp.Applying && diff < bestLastDiff {
			bestLastDiff = diff
			lastAsp = &HoraryAspect{Planet: name, Aspect: asp.Name, Orb: asp.Orb}
		}
	}

	return lastAsp, nextAsp
}
