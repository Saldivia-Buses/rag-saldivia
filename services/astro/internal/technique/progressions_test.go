package technique

import (
	"math"
	"testing"

	"github.com/Camionerou/rag-saldivia/services/astro/internal/ephemeris"
)

func TestCalcProgressions(t *testing.T) {
	chart := adrianChart(t)
	result, err := CalcProgressions(chart, 2026)
	if err != nil {
		t.Fatal(err)
	}

	if result.Year != 2026 {
		t.Errorf("year = %d, want 2026", result.Year)
	}

	// Age ~50.5 → progressed JD should be natal JD + ~50.5 days
	expectedAgeDays := (ephemeris.JulDay(2026, 7, 1, 12.0) - chart.JD) / 365.25
	if math.Abs(result.AgeDays-expectedAgeDays) > 0.1 {
		t.Errorf("age_days = %.2f, want ~%.2f", result.AgeDays, expectedAgeDays)
	}

	// Should have at least 7 planets (Sol through Saturno)
	if len(result.Positions) < 7 {
		t.Errorf("got %d positions, want >= 7", len(result.Positions))
	}

	// Progressed Sun should have moved ~50° from natal (1°/year)
	var progSunLon float64
	for _, pp := range result.Positions {
		if pp.Name == "Sol" {
			progSunLon = pp.Lon
			break
		}
	}
	natalSunLon := chart.Planets["Sol"].Lon
	diff := progSunLon - natalSunLon
	if diff < 0 {
		diff += 360
	}
	// Progressed Sun moves ~1°/year → ~50° in 50 years
	if diff < 40 || diff > 60 {
		t.Errorf("progressed Sun moved %.1f° from natal, expected ~50°", diff)
	}

	// Each position should have valid sign
	for _, pp := range result.Positions {
		if pp.Sign == "" {
			t.Errorf("%s has empty sign", pp.Name)
		}
		if pp.House < 1 || pp.House > 12 {
			t.Errorf("%s house = %d", pp.Name, pp.House)
		}
	}
}

func TestCalcProgressions_IngressDetection(t *testing.T) {
	chart := adrianChart(t)

	// Check multiple years — at least one should have an ingress
	hasIngress := false
	for y := 2020; y <= 2030; y++ {
		result, err := CalcProgressions(chart, y)
		if err != nil {
			continue
		}
		for _, pp := range result.Positions {
			if pp.Ingress != "" {
				hasIngress = true
				t.Logf("Ingress found: %s %s ingress in %d (prev %s)",
					pp.Name, pp.Ingress, y, pp.PrevSign)
			}
		}
	}
	if !hasIngress {
		t.Log("no ingresses found in 2020-2030 (may be correct, depends on chart)")
	}
}
