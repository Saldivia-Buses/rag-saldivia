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

// isVOC checks if the Moon is void-of-course (no applying major aspects before sign change).
// Correct method: for each planet, compute where each exact aspect angle falls on the ecliptic.
// If the Moon will reach any of those points before exiting the sign, it's NOT void.
// The Moon must be APPLYING (moving toward the aspect, not separating from it).
func isVOC(moonPos *ephemeris.PlanetPos, chart *natal.Chart) bool {
	moonLon := moonPos.Lon
	moonSign := astromath.SignIndex(moonLon)
	// Degrees from Moon's current position to the END of its current sign
	signEnd := float64(moonSign+1) * 30.0
	degsToSignEnd := signEnd - moonLon
	if degsToSignEnd <= 0 {
		degsToSignEnd += 360
	}
	if degsToSignEnd > 30 {
		degsToSignEnd = 30 // safety
	}

	aspectAngles := []float64{0, 60, 90, 120, 180}

	for name, pos := range chart.Planets {
		if name == "Luna" || name == "Nodo Norte" || name == "Nodo Sur" ||
			name == "Fortuna" || name == "Espíritu" {
			continue // skip non-physical points
		}
		planetLon := pos.Lon

		for _, aspAngle := range aspectAngles {
			// Two possible exact aspect points (planet + angle, planet - angle)
			for _, sign := range []float64{1, -1} {
				if aspAngle == 0 && sign == -1 { continue }
				if aspAngle == 180 && sign == -1 { continue }

				exactPoint := astromath.Normalize360(planetLon + aspAngle*sign)
				// How far does the Moon need to travel to reach this point?
				dist := astromath.Normalize360(exactPoint - moonLon)

				// Moon is applying if: distance is positive (ahead in zodiac) and within sign
				if dist > 0 && dist < degsToSignEnd {
					return false // Moon WILL form this aspect before sign change — NOT void
				}
			}
		}
	}

	return true // no applying aspects found — Moon is VOC
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
