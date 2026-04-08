package technique

import (
	"encoding/json"
	"math"
	"os"
	"testing"

	"github.com/Camionerou/rag-saldivia/services/astro/internal/ephemeris"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/natal"
)

func TestMain(m *testing.M) {
	p := os.Getenv("EPHE_PATH")
	ephemeris.Init(p)
	code := m.Run()
	ephemeris.Close()
	os.Exit(code)
}

// adrianChart builds Adrian's natal chart for test reuse.
func adrianChart(t *testing.T) *natal.Chart {
	t.Helper()
	chart, err := natal.BuildNatal(1975, 12, 27, 16.0+14.0/60.0, -32.9468, -60.6393, 25.0, -3)
	if err != nil {
		t.Fatal(err)
	}
	return chart
}

func TestCalcSolarArc_Golden(t *testing.T) {
	data, err := os.ReadFile("../../testdata/golden/solar_arc_adrian_2026.json")
	if err != nil {
		t.Skip("golden file not found")
	}

	var golden struct {
		Output struct {
			ArcDeg    float64                       `json:"arc_deg"`
			Positions map[string]struct {
				NatalLon float64 `json:"natal_lon"`
				SALon    float64 `json:"sa_lon"`
			} `json:"positions"`
		} `json:"output"`
	}
	if err := json.Unmarshal(data, &golden); err != nil {
		t.Fatal(err)
	}

	chart := adrianChart(t)
	sa := CalcSolarArcForYear(chart, 2026)

	// Verify arc matches Python (tolerance 0.05° — Go uses July 1, Python June 15 = ~16 day drift)
	if math.Abs(sa.ArcDeg-golden.Output.ArcDeg) > 0.05 {
		t.Errorf("arc = %.4f, Python = %.4f", sa.ArcDeg, golden.Output.ArcDeg)
	}

	// Verify SA positions match Python
	for name, pyPos := range golden.Output.Positions {
		goLon, ok := sa.Positions[name]
		if !ok {
			continue
		}
		if math.Abs(goLon-pyPos.SALon) > 0.05 {
			t.Errorf("%s: SA lon = %.4f, Python = %.4f", name, goLon, pyPos.SALon)
		}
	}
}

func TestFindSolarArcActivations(t *testing.T) {
	chart := adrianChart(t)
	jdMid := ephemeris.JulDay(2026, 6, 15, 12.0)
	results := FindSolarArcActivations(chart, jdMid)

	if len(results) == 0 {
		t.Error("expected at least some SA activations for 2026")
	}

	// Verify all results have valid fields
	for i, r := range results {
		if r.SAplanet == "" || r.NatPlanet == "" || r.Aspect == "" {
			t.Errorf("result[%d] has empty fields: %+v", i, r)
		}
		if r.Orb < 0 {
			t.Errorf("result[%d] has negative orb: %.4f", i, r.Orb)
		}
		if r.Nature != "fácil" && r.Nature != "tenso" && r.Nature != "neutral" {
			t.Errorf("result[%d] invalid nature: %q", i, r.Nature)
		}
	}
}

func TestAspectNature(t *testing.T) {
	if aspectNature("trine") != "fácil" {
		t.Error("trine should be fácil")
	}
	if aspectNature("square") != "tenso" {
		t.Error("square should be tenso")
	}
	if aspectNature("conjunction") != "neutral" {
		t.Error("conjunction should be neutral")
	}
}
