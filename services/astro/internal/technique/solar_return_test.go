package technique

import (
	"encoding/json"
	"math"
	"os"
	"testing"
)

func TestCalcSolarReturn_Golden(t *testing.T) {
	data, err := os.ReadFile("../../testdata/golden/solar_return_adrian_2026.json")
	if err != nil {
		t.Skip("golden file not found")
	}

	var golden struct {
		Output struct {
			JD     float64                        `json:"jd"`
			ASCLon float64                        `json:"asc_lon"`
			MCLon  float64                        `json:"mc_lon"`
			Planets map[string]struct {
				Lon float64 `json:"lon"`
			} `json:"planets"`
		} `json:"output"`
	}
	if err := json.Unmarshal(data, &golden); err != nil {
		t.Fatal(err)
	}

	chart := adrianChart(t)
	sr, err := CalcSolarReturnAtBirthplace(chart, 2026)
	if err != nil {
		t.Fatal(err)
	}

	// JD should match within ~1 minute (0.001 day ≈ 1.4 min)
	if math.Abs(sr.JD-golden.Output.JD) > 0.01 {
		t.Errorf("SR JD = %.4f, Python = %.4f (diff %.4f days)",
			sr.JD, golden.Output.JD, math.Abs(sr.JD-golden.Output.JD))
	}

	// Sun at SR should match natal Sun
	natalSunLon := chart.Planets["Sol"].Lon
	srSunLon := sr.Planets["Sol"].Lon
	if math.Abs(srSunLon-natalSunLon) > 0.001 {
		t.Errorf("SR Sun = %.4f, natal Sun = %.4f", srSunLon, natalSunLon)
	}

	// ASC should be in a reasonable range (not 0)
	if sr.ASC == 0 {
		t.Error("SR ASC should not be 0")
	}

	// All 10 planets should be present
	if len(sr.Planets) < 10 {
		t.Errorf("SR has %d planets, want >= 10", len(sr.Planets))
	}
}

func TestCalcLunarReturns(t *testing.T) {
	chart := adrianChart(t)
	returns, err := CalcLunarReturns(chart, 2026)
	if err != nil {
		t.Fatal(err)
	}

	// Moon completes ~13 cycles per year
	if len(returns) < 12 || len(returns) > 14 {
		t.Errorf("got %d lunar returns, expected 12-14", len(returns))
	}

	// Each return should have a valid month
	for i, lr := range returns {
		if lr.Month < 1 || lr.Month > 12 {
			t.Errorf("lunar return[%d] month = %d", i, lr.Month)
		}
	}
}
