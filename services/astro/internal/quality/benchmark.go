package quality

import "fmt"

// BenchmarkGap represents a gap between current performance and target.
type BenchmarkGap struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Target   int     `json:"target"`   // target percentage
	Current  int     `json:"current"`  // current percentage
	Delta    int     `json:"delta"`    // target - current (positive = gap)
	Weight   int     `json:"weight"`   // priority weight
	Priority float64 `json:"priority"` // delta × weight / 10
	Status   string  `json:"status"`   // "met", "close", "gap"
}

// BenchmarkReport holds the full benchmark analysis.
type BenchmarkReport struct {
	AvgScore    int            `json:"avg_score"`
	TargetScore int            `json:"target_score"` // 85
	EliteScore  int            `json:"elite_score"`  // 92
	Gaps        []BenchmarkGap `json:"gaps"`
	MetCount    int            `json:"met_count"`
	TotalChecks int            `json:"total_checks"`
}

// BenchmarkTargets defines the quality targets.
var BenchmarkTargets = []struct {
	ID     string
	Name   string
	Target int
	Weight int
}{
	{"no_hallucination", "Sin meses inventados", 100, 10},
	{"techniques_cited", "Técnicas mencionadas vs disponibles", 80, 8},
	{"convergence_used", "Convergencias incorporadas", 90, 9},
	{"precautions_met", "Precauciones cumplidas", 85, 7},
	{"actionable", "Recomendación accionable presente", 95, 8},
	{"length_adequate", "Largo de respuesta adecuado", 90, 5},
	{"structure", "Respuesta con estructura (headers)", 85, 4},
}

// AnalyzeBenchmark compares audit results against benchmark targets.
func AnalyzeBenchmark(auditResults []AuditResult) *BenchmarkReport {
	report := &BenchmarkReport{
		TargetScore: 85,
		EliteScore:  92,
		TotalChecks: len(BenchmarkTargets),
	}

	if len(auditResults) == 0 {
		return report
	}

	// Average score
	totalScore := 0
	for _, a := range auditResults {
		totalScore += a.ScoreTotal
	}
	report.AvgScore = totalScore / len(auditResults)

	// Analyze each benchmark target
	for _, target := range BenchmarkTargets {
		current := measureTarget(target.ID, auditResults)
		delta := target.Target - current
		priority := float64(max(0, delta)) * float64(target.Weight) / 10

		status := "gap"
		if delta <= 0 { status = "met" }
		if delta > 0 && delta <= 15 { status = "close" }

		gap := BenchmarkGap{
			ID: target.ID, Name: target.Name,
			Target: target.Target, Current: current,
			Delta: delta, Weight: target.Weight,
			Priority: priority, Status: status,
		}
		report.Gaps = append(report.Gaps, gap)
		if status == "met" { report.MetCount++ }
	}

	// Sort by priority descending
	for i := 0; i < len(report.Gaps); i++ {
		for j := i + 1; j < len(report.Gaps); j++ {
			if report.Gaps[j].Priority > report.Gaps[i].Priority {
				report.Gaps[i], report.Gaps[j] = report.Gaps[j], report.Gaps[i]
			}
		}
	}

	return report
}

// measureTarget computes the current % achievement for a benchmark target.
func measureTarget(targetID string, results []AuditResult) int {
	total := len(results)
	if total == 0 { return 0 }

	passed := 0
	for _, a := range results {
		switch targetID {
		case "no_hallucination":
			hasHallucination := false
			for _, issue := range a.Issues {
				if issue.Category == "hallucination" { hasHallucination = true; break }
			}
			if !hasHallucination { passed++ }
		case "techniques_cited":
			if a.TechniquesExpected > 0 && float64(a.TechniquesUsed)/float64(a.TechniquesExpected) >= 0.8 { passed++ }
		case "precautions_met":
			if a.PrecautionsTotal > 0 && float64(a.PrecautionsMet)/float64(a.PrecautionsTotal) >= 0.85 { passed++ }
		case "actionable":
			if a.ScoreCommunication >= 70 { passed++ }
		case "length_adequate":
			if a.ScoreCommunication >= 60 { passed++ }
		case "structure":
			if a.ScoreCommunication >= 80 { passed++ }
		case "convergence_used":
			if a.ScoreTechnical >= 70 { passed++ }
		}
	}

	return passed * 100 / total
}

// FormatBenchmarkReport generates a text report.
func FormatBenchmarkReport(report *BenchmarkReport) string {
	status := "🔴"
	if report.AvgScore >= report.EliteScore {
		status = "🟢 ELITE"
	} else if report.AvgScore >= report.TargetScore {
		status = "🟢"
	} else if report.AvgScore >= report.TargetScore-10 {
		status = "🟡"
	}

	result := fmt.Sprintf("Score promedio: %d/100 %s (target: %d, elite: %d)\nChecks: %d/%d met\n\n",
		report.AvgScore, status, report.TargetScore, report.EliteScore, report.MetCount, report.TotalChecks)

	for _, g := range report.Gaps {
		icon := "✅"
		if g.Status == "close" { icon = "🟡" }
		if g.Status == "gap" { icon = "🔴" }
		result += fmt.Sprintf("%s %s: %d%% (target %d%%, delta %d)\n", icon, g.Name, g.Current, g.Target, g.Delta)
	}

	return result
}
