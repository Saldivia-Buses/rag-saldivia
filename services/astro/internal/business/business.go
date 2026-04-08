// Package business provides astrological business intelligence.
// All calculations are based on a company's natal chart (kind="empresa")
// and its counterparties. Generic design — works for any business.
package business

import (
	"github.com/Camionerou/rag-saldivia/services/astro/internal/natal"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/technique"
)

// Service provides business intelligence calculations.
type Service struct{}

// NewService creates a business intelligence service.
func NewService() *Service {
	return &Service{}
}

// BuildDashboard produces the full business dashboard for a company chart.
func (s *Service) BuildDashboard(
	companyChart *natal.Chart,
	companyName string,
	counterpartyCharts map[string]*natal.Chart, // name → chart
	year, month int,
) *DashboardResponse {
	resp := &DashboardResponse{
		CompanyName: companyName,
		Year:        year,
		Month:       month,
	}

	// Cash flow (full year)
	resp.CashFlow = CalcCashFlow(companyChart, year)

	// Risk heatmap (full year)
	resp.RiskHeatmap = CalcRiskHeatmap(companyChart, year)

	// Timing windows (for each counterparty)
	for name, chart := range counterpartyCharts {
		windows := CalcNegotiationTiming(companyChart, chart, name, year, month)
		resp.TopTimings = append(resp.TopTimings, windows...)
	}

	// Mercury Rx calendar
	resp.MercuryRx = CalcMercuryRx(year)

	// Quarterly forecast
	quarter := (month-1)/3 + 1
	resp.Forecast = CalcQuarterlyForecast(companyChart, year, quarter)

	// Today's agenda
	resp.TodayAgenda = BuildAgenda(companyChart, resp.TopTimings, resp.RiskHeatmap, year, month, 0) // day=0 means "today-like"

	// Week highlights
	resp.WeekHighlights = buildWeekHighlights(resp)

	return resp
}

// buildWeekHighlights extracts the top 3-5 notable items from the dashboard.
func buildWeekHighlights(resp *DashboardResponse) []string {
	var highlights []string

	// High-score timing windows this month
	for _, tw := range resp.TopTimings {
		if tw.Month == resp.Month && tw.Score >= 70 {
			highlights = append(highlights, "Ventana favorable para "+tw.Counterparty)
		}
	}

	// High risk months
	for _, rc := range resp.RiskHeatmap {
		if rc.Month == resp.Month && rc.Level >= 4 {
			highlights = append(highlights, "Riesgo "+rc.Category+": "+rc.Alert)
		}
	}

	// Mercury Rx
	for _, rx := range resp.MercuryRx {
		if rx.StartMonth == resp.Month || rx.EndMonth == resp.Month {
			highlights = append(highlights, "Mercurio retrógrado en "+rx.Sign+" — revisar contratos")
		}
	}

	// Cash flow
	for _, cf := range resp.CashFlow {
		if cf.Month == resp.Month {
			if cf.Rating == "presión" {
				highlights = append(highlights, "Cash flow bajo presión — "+cf.Details)
			} else if cf.Rating == "abundancia" {
				highlights = append(highlights, "Cash flow favorable — "+cf.Details)
			}
		}
	}

	if len(highlights) > 5 {
		highlights = highlights[:5]
	}
	return highlights
}

// TeamCompatibility computes synastry scores between company and team members.
func (s *Service) TeamCompatibility(
	companyChart *natal.Chart,
	memberCharts map[string]*natal.Chart,
) []TeamScore {
	var scores []TeamScore

	for name, memberChart := range memberCharts {
		pair := &technique.ChartPair{
			ChartA: companyChart,
			ChartB: memberChart,
			NameA:  "Empresa",
			NameB:  name,
		}
		syn := technique.CalcSynastry(pair)

		ts := TeamScore{
			Name:  name,
			Score: syn.Score,
		}
		for _, c := range syn.Connections {
			ts.Strengths = append(ts.Strengths, c.PlanetA+" "+c.Aspect+" "+c.PlanetB)
		}
		for _, f := range syn.Frictions {
			ts.Tensions = append(ts.Tensions, f.PlanetA+" "+f.Aspect+" "+f.PlanetB)
		}
		if len(ts.Strengths) > 3 {
			ts.Strengths = ts.Strengths[:3]
		}
		if len(ts.Tensions) > 3 {
			ts.Tensions = ts.Tensions[:3]
		}
		scores = append(scores, ts)
	}

	return scores
}
