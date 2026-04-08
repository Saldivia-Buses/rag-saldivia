package business

import (
	"github.com/Camionerou/rag-saldivia/services/astro/internal/astromath"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/ephemeris"
)

// CalcMercuryRx finds all Mercury retrograde periods in a year.
// Scans daily to detect when Mercury's speed goes negative.
func CalcMercuryRx(year int) []MercuryRxPeriod {
	flags := ephemeris.FlagSwieph | ephemeris.FlagSpeed
	jdStart := ephemeris.JulDay(year, 1, 1, 0.0)
	jdEnd := ephemeris.JulDay(year+1, 1, 1, 0.0)

	var periods []MercuryRxPeriod
	inRx := false
	var current MercuryRxPeriod

	for jd := jdStart; jd < jdEnd; jd += 1.0 {
		pos, err := ephemeris.CalcPlanet(jd, ephemeris.Mercury, flags)
		if err != nil {
			continue
		}

		if pos.Speed < 0 && !inRx {
			// Start of Rx period
			inRx = true
			_, m, d, _ := ephemeris.RevJul(jd)
			current = MercuryRxPeriod{
				StartMonth: m,
				StartDay:   d,
				Sign:       astromath.SignName(pos.Lon),
			}
		} else if pos.Speed >= 0 && inRx {
			// End of Rx period
			inRx = false
			_, m, d, _ := ephemeris.RevJul(jd)
			current.EndMonth = m
			current.EndDay = d
			current.Impact = mercuryRxImpact(current.Sign)
			periods = append(periods, current)
		}
	}

	// Handle Rx that extends past year end
	if inRx {
		current.EndMonth = 12
		current.EndDay = 31
		current.Impact = mercuryRxImpact(current.Sign)
		periods = append(periods, current)
	}

	return periods
}

func mercuryRxImpact(sign string) string {
	switch sign {
	case "Aries", "Leo", "Sagitario":
		return "Afecta decisiones ejecutivas y liderazgo. Revisar estrategia."
	case "Tauro", "Virgo", "Capricornio":
		return "Afecta contratos, finanzas y operaciones. No firmar sin revisar."
	case "Géminis", "Libra", "Acuario":
		return "Afecta comunicaciones y negociaciones. Confirmar todo por escrito."
	case "Cáncer", "Escorpio", "Piscis":
		return "Afecta relaciones internas y decisiones emocionales. Prudencia."
	default:
		return "Revisar documentos y comunicaciones importantes."
	}
}
