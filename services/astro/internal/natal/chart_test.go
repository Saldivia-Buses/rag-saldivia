package natal

import (
	"encoding/json"
	"math"
	"os"
	"testing"

	"github.com/Camionerou/rag-saldivia/services/astro/internal/astromath"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/ephemeris"
)

func TestMain(m *testing.M) {
	p := os.Getenv("EPHE_PATH")
	ephemeris.Init(p)
	code := m.Run()
	ephemeris.Close()
	os.Exit(code)
}

type goldenPlanet struct {
	RA    float64 `json:"ra"`
	Dec   float64 `json:"dec"`
	Lon   float64 `json:"lon"`
	Lat   float64 `json:"lat"`
	Speed float64 `json:"speed"`
}

type goldenNatal struct {
	Input struct {
		Year      int     `json:"year"`
		Month     int     `json:"month"`
		Day       int     `json:"day"`
		Hour      int     `json:"hour"`
		Minute    int     `json:"minute"`
		Lat       float64 `json:"lat"`
		Lon       float64 `json:"lon"`
		Alt       float64 `json:"alt"`
		UTCOffset int     `json:"utc_offset"`
	} `json:"input"`
	Output struct {
		JD      float64                     `json:"jd"`
		Eps     float64                     `json:"eps"`
		RAMC    float64                     `json:"ramc"`
		Cusps   []float64                   `json:"cusps"`
		Planets map[string]*goldenPlanet    `json:"planets"`
	} `json:"output"`
}

func TestBuildNatal_Golden(t *testing.T) {
	data, err := os.ReadFile("../../testdata/golden/natal_adrian.json")
	if err != nil {
		t.Skip("golden file not found — run generate_golden.py first")
	}
	var golden goldenNatal
	if err := json.Unmarshal(data, &golden); err != nil {
		t.Fatal("unmarshal golden:", err)
	}

	in := golden.Input
	chart, err := BuildNatal(
		in.Year, in.Month, in.Day,
		float64(in.Hour)+float64(in.Minute)/60.0,
		in.Lat, in.Lon, in.Alt, in.UTCOffset,
	)
	if err != nil {
		t.Fatal(err)
	}

	// Verify JD
	if math.Abs(chart.JD-golden.Output.JD) > 0.001 {
		t.Errorf("JD = %.6f, Python = %.6f", chart.JD, golden.Output.JD)
	}

	// Verify epsilon
	if math.Abs(chart.Epsilon-golden.Output.Eps) > 0.01 {
		t.Errorf("Epsilon = %.6f, Python = %.6f", chart.Epsilon, golden.Output.Eps)
	}

	// Verify planet ecliptic longitudes match Python within 0.01°
	// (Moshier vs Swiss Ephemeris may differ slightly)
	for name, pyData := range golden.Output.Planets {
		goPos, ok := chart.Planets[name]
		if !ok {
			// Skip planets that Go doesn't calculate (Ecl.Prenatal, Vértice, AS, MC)
			continue
		}
		if math.Abs(goPos.Lon-pyData.Lon) > 0.01 {
			t.Errorf("%s: lon = %.4f, Python = %.4f (diff %.4f)",
				name, goPos.Lon, pyData.Lon, math.Abs(goPos.Lon-pyData.Lon))
		}
		// Skip RA/Dec for calculated points (no real equatorial position)
		calcPoints := map[string]bool{"Fortuna": true, "Espíritu": true, "Nodo Sur": true, "Lilith": true, "AS": true, "MC": true, "Vértice": true}
		if !calcPoints[name] && pyData.RA != 0 && goPos.RA != 0 {
			if math.Abs(goPos.RA-pyData.RA) > 0.05 {
				t.Errorf("%s: RA = %.4f, Python = %.4f", name, goPos.RA, pyData.RA)
			}
			if math.Abs(goPos.Dec-pyData.Dec) > 0.05 {
				t.Errorf("%s: Dec = %.4f, Python = %.4f", name, goPos.Dec, pyData.Dec)
			}
		}
	}

	// Verify South Node is 180° from North Node
	nn, ok1 := chart.Planets["Nodo Norte"]
	sn, ok2 := chart.Planets["Nodo Sur"]
	if ok1 && ok2 {
		diff := math.Abs(sn.Lon - astromath.Normalize360(nn.Lon+180))
		if diff > 0.001 {
			t.Errorf("South Node not 180° from North Node: SN=%.2f, NN=%.2f", sn.Lon, nn.Lon)
		}
	}

	// Verify Part of Fortune exists
	if _, ok := chart.Planets["Fortuna"]; !ok {
		t.Error("missing Part of Fortune")
	}

	// Verify Part of Spirit exists
	if _, ok := chart.Planets["Espíritu"]; !ok {
		t.Error("missing Part of Spirit")
	}

	// Verify diurnal determination
	// Adrian born 16:14 local, Dec 27 in Rosario — Sun above horizon = diurnal
	if !chart.Diurnal {
		t.Error("expected diurnal chart (Sun above horizon at 16:14 Dec 27 Rosario)")
	}

	// Verify combustion map populated
	if len(chart.Combustion) == 0 {
		t.Error("combustion map empty")
	}
}

func TestBuildNatal_BasicFields(t *testing.T) {
	chart, err := BuildNatal(2000, 1, 1, 12.0, 0, 0, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if chart.JD == 0 {
		t.Error("JD should not be 0")
	}
	if chart.ASC == 0 && chart.MC == 0 {
		t.Error("ASC and MC should not both be 0")
	}
	if len(chart.Planets) < 10 {
		t.Errorf("expected at least 10 planets, got %d", len(chart.Planets))
	}
}
