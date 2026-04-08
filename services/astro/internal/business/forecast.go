package business

import (
	"fmt"

	"github.com/Camionerou/rag-saldivia/services/astro/internal/astromath"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/ephemeris"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/natal"
)

// CalcQuarterlyForecast produces a 3-month business outlook.
func CalcQuarterlyForecast(companyChart *natal.Chart, year, quarter int) *QuarterlyForecast {
	startMonth := (quarter-1)*3 + 1
	flags := ephemeris.FlagSwieph | ephemeris.FlagSpeed

	forecast := &QuarterlyForecast{
		Quarter: quarter,
		Year:    year,
	}

	positiveCount := 0
	challengeCount := 0

	// Check Jupiter and Saturn positions for the quarter midpoint
	midMonth := startMonth + 1
	jd := ephemeris.JulDay(year, midMonth, 15, 12.0)

	// Jupiter — expansion indicator
	jupPos, err := ephemeris.CalcPlanet(jd, ephemeris.Jupiter, flags)
	if err == nil {
		jupHouse := astromath.HouseForLon(jupPos.Lon, companyChart.Cusps)
		switch jupHouse {
		case 1, 10, 2:
			forecast.KeyEvents = append(forecast.KeyEvents,
				fmt.Sprintf("Júpiter en casa %d — período de expansión y oportunidades", jupHouse))
			positiveCount += 2
		case 7:
			forecast.KeyEvents = append(forecast.KeyEvents,
				"Júpiter en casa 7 — asociaciones favorables")
			positiveCount++
		}

		// Jupiter aspects to MC
		asp := astromath.FindAspect(jupPos.Lon, companyChart.MC, 5.0)
		if asp != nil && (asp.Name == "trine" || asp.Name == "conjunction") {
			forecast.ActionItems = append(forecast.ActionItems,
				"Aprovechar Júpiter favorable al MC para iniciativas de crecimiento")
			positiveCount++
		}
	}

	// Saturn — restriction/restructuring indicator
	satPos, err := ephemeris.CalcPlanet(jd, ephemeris.Saturn, flags)
	if err == nil {
		satHouse := astromath.HouseForLon(satPos.Lon, companyChart.Cusps)
		switch satHouse {
		case 10:
			forecast.KeyEvents = append(forecast.KeyEvents,
				"Saturno en casa 10 — reestructuración de autoridad y procesos")
			challengeCount++
			forecast.ActionItems = append(forecast.ActionItems,
				"Revisar estructura organizacional y roles de liderazgo")
		case 2:
			forecast.KeyEvents = append(forecast.KeyEvents,
				"Saturno en casa 2 — presión sobre recursos, necesidad de austeridad")
			challengeCount += 2
			forecast.ActionItems = append(forecast.ActionItems,
				"Ajustar presupuesto y priorizar gastos esenciales")
		case 8:
			forecast.KeyEvents = append(forecast.KeyEvents,
				"Saturno en casa 8 — negociaciones de deuda, restructuración financiera")
			challengeCount++
		}
	}

	// Mercury Rx during this quarter
	for m := startMonth; m < startMonth+3 && m <= 12; m++ {
		jdm := ephemeris.JulDay(year, m, 15, 12.0)
		mercPos, err := ephemeris.CalcPlanet(jdm, ephemeris.Mercury, flags)
		if err == nil && mercPos.Speed < 0 {
			forecast.KeyEvents = append(forecast.KeyEvents,
				fmt.Sprintf("Mercurio retrógrado en mes %d — revisar contratos antes de firmar", m))
			forecast.ActionItems = append(forecast.ActionItems,
				"No firmar contratos nuevos durante Mercurio retrógrado sin doble revisión")
			challengeCount++
		}
	}

	// Determine outlook
	switch {
	case positiveCount > challengeCount+1:
		forecast.Outlook = "positivo"
		forecast.Summary = "Trimestre favorable para la expansión. Las oportunidades superan los desafíos."
	case challengeCount > positiveCount+1:
		forecast.Outlook = "desafiante"
		forecast.Summary = "Trimestre que requiere cautela. Enfocarse en consolidación y gestión de riesgos."
	default:
		forecast.Outlook = "neutro"
		forecast.Summary = "Trimestre mixto. Oportunidades disponibles pero con desafíos que gestionar."
	}

	// Default action items if none generated
	if len(forecast.ActionItems) == 0 {
		forecast.ActionItems = []string{"Mantener operaciones estables", "Monitorear indicadores clave"}
	}

	return forecast
}
