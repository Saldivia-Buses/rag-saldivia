package business

import (
	"github.com/Camionerou/rag-saldivia/services/astro/internal/natal"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/technique"
)

// SuccessionPlan holds a succession planning analysis.
type SuccessionPlan struct {
	Candidates []SuccessionCandidate `json:"candidates"`
	BestWindow string                `json:"best_window"` // timing recommendation
}

// SuccessionCandidate evaluates a potential successor.
type SuccessionCandidate struct {
	Name       string `json:"name"`
	Score      int    `json:"score"`
	LeaderFit  string `json:"leader_fit"`  // "natural", "desarrollable", "bajo"
	Strengths  []string `json:"strengths"`
}

// AnalyzeSuccession evaluates candidates for business succession/transition.
func AnalyzeSuccession(companyChart *natal.Chart, candidates map[string]*natal.Chart) *SuccessionPlan {
	plan := &SuccessionPlan{}

	for name, chart := range candidates {
		pair := &technique.ChartPair{
			ChartA: companyChart, ChartB: chart,
			NameA: "Empresa", NameB: name,
		}
		syn := technique.CalcSynastry(pair)

		fit := "desarrollable"
		if syn.Score >= 70 { fit = "natural" }
		if syn.Score < 40 { fit = "bajo" }

		var strengths []string
		for _, c := range syn.Connections {
			strengths = append(strengths, c.PlanetA+" "+c.Aspect+" "+c.PlanetB)
			if len(strengths) >= 3 { break }
		}

		plan.Candidates = append(plan.Candidates, SuccessionCandidate{
			Name: name, Score: syn.Score, LeaderFit: fit, Strengths: strengths,
		})
	}

	// Sort by score
	for i := 0; i < len(plan.Candidates); i++ {
		for j := i + 1; j < len(plan.Candidates); j++ {
			if plan.Candidates[j].Score > plan.Candidates[i].Score {
				plan.Candidates[i], plan.Candidates[j] = plan.Candidates[j], plan.Candidates[i]
			}
		}
	}

	plan.BestWindow = "Evaluar timing con técnicas predictivas de la empresa"
	return plan
}
