package business

import (
	"github.com/Camionerou/rag-saldivia/services/astro/internal/astromath"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/natal"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/technique"
)

// EmployeeScreening holds an astrological compatibility assessment for hiring.
type EmployeeScreening struct {
	CandidateName string   `json:"candidate_name"`
	Score         int      `json:"score"`      // 0-100
	Strengths     []string `json:"strengths"`
	Concerns      []string `json:"concerns"`
	BestRole      string   `json:"best_role"`  // suggested role alignment
}

// ScreenEmployee evaluates a candidate's chart against the company chart.
func ScreenEmployee(companyChart, candidateChart *natal.Chart, candidateName string) *EmployeeScreening {
	pair := &technique.ChartPair{
		ChartA: companyChart,
		ChartB: candidateChart,
		NameA:  "Empresa",
		NameB:  candidateName,
	}
	syn := technique.CalcSynastry(pair)

	result := &EmployeeScreening{
		CandidateName: candidateName,
		Score:         syn.Score,
	}

	for _, c := range syn.Connections {
		result.Strengths = append(result.Strengths, c.PlanetA+" "+c.Aspect+" "+c.PlanetB)
	}
	for _, f := range syn.Frictions {
		result.Concerns = append(result.Concerns, f.PlanetA+" "+f.Aspect+" "+f.PlanetB)
	}

	// Suggest role based on candidate's strongest house
	if p, ok := candidateChart.Planets["Sol"]; ok {
		house := astromath.HouseForLon(p.Lon, candidateChart.Cusps)
		switch house {
		case 1, 10:
			result.BestRole = "liderazgo"
		case 2, 8:
			result.BestRole = "finanzas"
		case 3, 9:
			result.BestRole = "comunicación/ventas"
		case 6:
			result.BestRole = "operaciones"
		case 7:
			result.BestRole = "relaciones/RRHH"
		default:
			result.BestRole = "general"
		}
	}

	if len(result.Strengths) > 3 { result.Strengths = result.Strengths[:3] }
	if len(result.Concerns) > 3 { result.Concerns = result.Concerns[:3] }

	return result
}
