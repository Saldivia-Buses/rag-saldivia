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

	expectedAge := (ephemeris.JulDay(2026, 7, 1, 12.0) - chart.JD) / 365.25
	if math.Abs(result.AgeYears-expectedAge) > 0.1 {
		t.Errorf("age_years = %.2f, want ~%.2f", result.AgeYears, expectedAge)
	}

	if len(result.Positions) < 7 {
		t.Errorf("got %d positions, want >= 7", len(result.Positions))
	}

	// Deterministic order: first should be Sol
	if result.Positions[0].Name != "Sol" {
		t.Errorf("first position = %q, want Sol", result.Positions[0].Name)
	}

	// Progressed Sun should have moved ~50° from natal
	progSunLon := result.Positions[0].Lon
	natalSunLon := chart.Planets["Sol"].Lon
	diff := progSunLon - natalSunLon
	if diff < 0 {
		diff += 360
	}
	if diff < 40 || diff > 60 {
		t.Errorf("progressed Sun moved %.1f° from natal, expected ~50°", diff)
	}

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

	signIngresses := 0
	houseIngresses := 0
	for y := 2020; y <= 2030; y++ {
		result, err := CalcProgressions(chart, y)
		if err != nil {
			continue
		}
		for _, pp := range result.Positions {
			if pp.SignIngress {
				signIngresses++
				t.Logf("%d: %s sign ingress (%s → %s)", y, pp.Name, pp.PrevSign, pp.Sign)
			}
			if pp.HouseIngress {
				houseIngresses++
			}
		}
	}
	if signIngresses == 0 {
		t.Error("no sign ingresses found in 2020-2030")
	}
	t.Logf("Total: %d sign ingresses, %d house ingresses", signIngresses, houseIngresses)
}
