package context

import (
	"encoding/json"
	"math"
	"os"
	"testing"
	"time"

	"github.com/Camionerou/rag-saldivia/services/astro/internal/astromath"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/ephemeris"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/natal"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/technique"
)

// goldenData matches the Python go_comparison.json structure.
type goldenData struct {
	Natal struct {
		Planets map[string]float64 `json:"planets"`
		JD      float64            `json:"jd"`
	} `json:"natal"`
	Profection struct {
		CasaActiva int     `json:"casa_activa"`
		Lord       string  `json:"lord"`
		ProfLon    float64 `json:"prof_lon"`
	} `json:"profection"`
	Firdaria struct {
		MajorLord string `json:"major_lord"`
		SubLord   string `json:"sub_lord"`
		Diurnal   bool   `json:"diurnal"`
	} `json:"firdaria"`
	Almuten struct {
		Winner  string `json:"winner"`
		Score   int    `json:"score"`
		Diurnal bool   `json:"diurnal"`
	} `json:"almuten"`
	Lots map[string]struct {
		Degree float64 `json:"degree"`
		Sign   string  `json:"sign"`
	} `json:"lots"`
	SolarReturn struct {
		JD float64 `json:"jd"`
	} `json:"solar_return"`
	Decennials struct {
		MajorPlanet string `json:"major_planet"`
		SubPlanet   string `json:"sub_planet"`
		Diurnal     bool   `json:"diurnal"`
	} `json:"decennials"`
	SabianSun struct {
		Lon    float64 `json:"lon"`
		Symbol string  `json:"symbol"`
	} `json:"sabian_sun"`
}

func loadGolden(t *testing.T) *goldenData {
	data, err := os.ReadFile("../../testdata/golden/python_comparison.json")
	if err != nil {
		t.Fatalf("load golden: %v", err)
	}
	var g goldenData
	if err := json.Unmarshal(data, &g); err != nil {
		t.Fatalf("parse golden: %v", err)
	}
	return &g
}

func adrianChartForComparison(t *testing.T) (*natal.Chart, time.Time) {
	t.Helper()
	ephePath := os.Getenv("EPHE_PATH")
	if ephePath == "" { ephePath = "/ephe" }
	ephemeris.Init(ephePath)

	chart, err := natal.BuildNatal(1975, 12, 27, 16.0+14.0/60.0, -32.9468, -60.6393, 25.0, -3)
	if err != nil {
		t.Fatalf("BuildNatal: %v", err)
	}
	birthDate := time.Date(1975, 12, 27, 0, 0, 0, 0, time.UTC)
	return chart, birthDate
}

// TestPythonComparison_Profection compares Go profection against Python.
func TestPythonComparison_Profection(t *testing.T) {
	golden := loadGolden(t)
	chart, birthDate := adrianChartForComparison(t)

	prof := technique.CalcProfection(chart, birthDate, 2026)
	if prof.ActiveHouse != golden.Profection.CasaActiva {
		t.Errorf("Profection house: Go=%d, Python=%d", prof.ActiveHouse, golden.Profection.CasaActiva)
	} else {
		t.Logf("✓ Profection house: %d", prof.ActiveHouse)
	}
	if prof.Lord != golden.Profection.Lord {
		t.Errorf("Profection lord: Go=%s, Python=%s", prof.Lord, golden.Profection.Lord)
	} else {
		t.Logf("✓ Profection lord: %s", prof.Lord)
	}
}

// TestPythonComparison_Firdaria compares Go firdaria against Python.
func TestPythonComparison_Firdaria(t *testing.T) {
	golden := loadGolden(t)
	chart, birthDate := adrianChartForComparison(t)

	fird := technique.CalcFirdaria(birthDate, chart.Diurnal, 2026)
	if fird.MajorLord != golden.Firdaria.MajorLord {
		t.Errorf("Firdaria major: Go=%s, Python=%s", fird.MajorLord, golden.Firdaria.MajorLord)
	} else {
		t.Logf("✓ Firdaria major: %s", fird.MajorLord)
	}
	if fird.SubLord != golden.Firdaria.SubLord {
		t.Errorf("Firdaria sub: Go=%s, Python=%s", fird.SubLord, golden.Firdaria.SubLord)
	} else {
		t.Logf("✓ Firdaria sub: %s", fird.SubLord)
	}
	if fird.Diurnal != golden.Firdaria.Diurnal {
		t.Errorf("Firdaria diurnal: Go=%v, Python=%v", fird.Diurnal, golden.Firdaria.Diurnal)
	}
}

// TestPythonComparison_Almuten compares Go almuten against Python.
func TestPythonComparison_Almuten(t *testing.T) {
	golden := loadGolden(t)
	chart, _ := adrianChartForComparison(t)

	alm := astromath.CalcAlmuten(chart.Planets, chart.ASC, chart.MC, chart.Diurnal)
	if alm.Winner != golden.Almuten.Winner {
		t.Errorf("Almuten winner: Go=%s, Python=%s", alm.Winner, golden.Almuten.Winner)
	} else {
		t.Logf("✓ Almuten winner: %s", alm.Winner)
	}
	if alm.Score != golden.Almuten.Score {
		t.Errorf("Almuten score: Go=%d, Python=%d", alm.Score, golden.Almuten.Score)
	} else {
		t.Logf("✓ Almuten score: %d", alm.Score)
	}
}

// TestPythonComparison_Lots compares Go hellenistic lots against Python.
func TestPythonComparison_Lots(t *testing.T) {
	golden := loadGolden(t)
	chart, _ := adrianChartForComparison(t)

	lots := astromath.CalcAllLots(chart.Planets, chart.ASC, chart.Diurnal, chart.Cusps)

	for _, lot := range lots {
		pyLot, ok := golden.Lots[lot.Name]
		if !ok {
			continue // Python may use different names
		}
		diff := math.Abs(lot.Lon - pyLot.Degree)
		if diff > 180 { diff = 360 - diff }
		if diff > 0.1 {
			t.Errorf("Lot %s: Go=%.4f°, Python=%.4f°, diff=%.4f°", lot.Name, lot.Lon, pyLot.Degree, diff)
		} else {
			t.Logf("✓ Lot %s: diff=%.4f°", lot.Name, diff)
		}
	}
}

// TestPythonComparison_SolarReturn compares Go solar return JD against Python.
func TestPythonComparison_SolarReturn(t *testing.T) {
	golden := loadGolden(t)
	chart, _ := adrianChartForComparison(t)

	sr, err := technique.CalcSolarReturnAtBirthplace(chart, 2026)
	if err != nil {
		t.Fatalf("Solar return: %v", err)
	}
	diff := math.Abs(sr.JD - golden.SolarReturn.JD)
	if diff > 0.001 { // 0.001 JD = ~1.4 minutes
		t.Errorf("Solar Return JD: Go=%.6f, Python=%.6f, diff=%.6f (%.1f min)", sr.JD, golden.SolarReturn.JD, diff, diff*1440)
	} else {
		t.Logf("✓ Solar Return JD: diff=%.6f (%.1f seconds)", diff, diff*86400)
	}
}

// TestPythonComparison_Decennials compares Go decennials against Python.
func TestPythonComparison_Decennials(t *testing.T) {
	golden := loadGolden(t)
	chart, birthDate := adrianChartForComparison(t)

	dec := technique.CalcDecennials(chart, birthDate, 2026)
	if dec.MajorPlanet != golden.Decennials.MajorPlanet {
		t.Errorf("Decennial major: Go=%s, Python=%s", dec.MajorPlanet, golden.Decennials.MajorPlanet)
	} else {
		t.Logf("✓ Decennial major: %s", dec.MajorPlanet)
	}
	if dec.SubPlanet != golden.Decennials.SubPlanet {
		t.Errorf("Decennial sub: Go=%s, Python=%s", dec.SubPlanet, golden.Decennials.SubPlanet)
	} else {
		t.Logf("✓ Decennial sub: %s", dec.SubPlanet)
	}
}

// TestPythonComparison_Sabian compares Go sabian symbol against Python.
func TestPythonComparison_Sabian(t *testing.T) {
	golden := loadGolden(t)

	symbol := astromath.SabianSymbol(golden.SabianSun.Lon)
	if symbol != golden.SabianSun.Symbol {
		t.Errorf("Sabian Sun:\n  Go:     %s\n  Python: %s", symbol, golden.SabianSun.Symbol)
	} else {
		t.Logf("✓ Sabian Sun: %s", symbol[:50])
	}
}

// TestPythonComparison_FullBuild verifies the full Build() runs without error.
func TestPythonComparison_FullBuild(t *testing.T) {
	chart, birthDate := adrianChartForComparison(t)

	ctx, err := Build(chart, "Adrian Saldivia", birthDate, 2026)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if ctx.Score == 0 {
		t.Error("Score should be > 0 for an active year")
	}
	t.Logf("✓ Full Build: score=%d, brief=%d chars, %d verdicts, %d contradictions",
		ctx.Score, len(ctx.Brief), len(ctx.Verdicts), len(ctx.Contradictions))

	// Verify all new Plan 12 fields are populated
	if ctx.Almuten == nil { t.Error("Almuten not computed") }
	if len(ctx.Lots) == 0 { t.Error("Lots not computed") }
	if ctx.Disposition == nil { t.Error("Disposition not computed") }
	if ctx.Sect == nil { t.Error("Sect not computed") }
	if ctx.Temperament == nil { t.Error("Temperament not computed") }
	if ctx.Divisor == nil { t.Error("Divisor not computed") }
	if len(ctx.AspectPatterns) == 0 { t.Log("No aspect patterns (may be normal)") }
	if ctx.ChartShape == nil { t.Error("ChartShape not computed") }
	if ctx.Hemispheres == nil { t.Error("Hemispheres not computed") }
	if len(ctx.FullDignities) == 0 { t.Error("FullDignities not computed") }
	if ctx.PlanetaryAge == nil { t.Error("PlanetaryAge not computed") }
	if len(ctx.Verdicts) == 0 { t.Error("Verdicts not computed") }
}
