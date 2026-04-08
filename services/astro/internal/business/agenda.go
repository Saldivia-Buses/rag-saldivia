package business

import (
	"fmt"

	"github.com/Camionerou/rag-saldivia/services/astro/internal/natal"
)

// BuildAgenda creates a daily agenda from timing windows, risks, and other sources.
// day=0 means build for the whole month overview.
func BuildAgenda(
	companyChart *natal.Chart,
	timings []TimingWindow,
	risks []RiskCell,
	year, month, day int,
) []AgendaItem {
	var items []AgendaItem

	// Timing windows for this month
	for _, tw := range timings {
		if tw.Month != month {
			continue
		}
		if day > 0 && (day < tw.DayStart || day > tw.DayEnd) {
			continue
		}
		dateStr := fmt.Sprintf("%d-%02d-%02d", year, month, tw.DayStart)
		if tw.DayStart != tw.DayEnd {
			dateStr = fmt.Sprintf("%d-%02d-%02d al %02d", year, month, tw.DayStart, tw.DayEnd)
		}
		items = append(items, AgendaItem{
			Date:        dateStr,
			Title:       "Ventana: " + tw.Counterparty,
			Description: tw.Nature + " — " + joinFactors(tw.Factors),
			Score:       tw.Score,
			Category:    "negotiation",
			Source:      "timing",
		})
	}

	// High-risk alerts for this month
	for _, rc := range risks {
		if rc.Month != month || rc.Level < 3 {
			continue
		}
		items = append(items, AgendaItem{
			Date:        fmt.Sprintf("%d-%02d-01", year, month),
			Title:       "Alerta: riesgo " + rc.Category,
			Description: rc.Alert,
			Score:       float64(rc.Level) * 20,
			Category:    "alert",
			Source:      "risk",
		})
	}

	// Sort by score descending
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			if items[j].Score > items[i].Score {
				items[i], items[j] = items[j], items[i]
			}
		}
	}

	return items
}

// CalcHiringCalendar finds favorable days for hiring based on H6 (employees) transits.
func CalcHiringCalendar(companyChart *natal.Chart, year, month int) []AgendaItem {
	// Simplified: favorable when Jupiter/Venus transit H6 cusp
	// This delegates to the timing engine with H6 focus
	// For now, returns the month's timing windows filtered for HR context
	var items []AgendaItem

	// H6 cusp = employees
	if len(companyChart.Cusps) <= 6 {
		return items
	}

	// Return placeholder — full implementation uses the timing engine
	items = append(items, AgendaItem{
		Date:     fmt.Sprintf("%d-%02d-01", year, month),
		Title:    "Ventana de contratación",
		Category: "hiring",
		Source:   "transit",
		Score:    50,
	})

	return items
}

func joinFactors(factors []string) string {
	if len(factors) == 0 {
		return ""
	}
	result := factors[0]
	for i := 1; i < len(factors) && i < 3; i++ {
		result += "; " + factors[i]
	}
	return result
}
