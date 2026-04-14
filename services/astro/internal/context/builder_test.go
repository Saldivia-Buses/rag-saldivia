package context

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Camionerou/rag-saldivia/services/astro/internal/ephemeris"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/natal"
)

func TestMain(m *testing.M) {
	p := os.Getenv("EPHE_PATH")
	if p == "" {
		// Ephemeris data not available — skip this package's tests (CGO/Linux-only).
		os.Exit(0)
	}
	ephemeris.Init(p)
	code := m.Run()
	ephemeris.Close()
	os.Exit(code)
}

func adrianChart(t *testing.T) *natal.Chart {
	t.Helper()
	chart, err := natal.BuildNatal(1975, 12, 27, 16.0+14.0/60.0, -32.9468, -60.6393, 25.0, -3)
	if err != nil {
		t.Fatal(err)
	}
	return chart
}

func TestBuild(t *testing.T) {
	chart := adrianChart(t)
	birthDate := time.Date(1975, 12, 27, 0, 0, 0, 0, time.UTC)

	ctx, err := Build(chart, "Adrian Saldivia", birthDate, 2026)
	if err != nil {
		t.Fatal(err)
	}

	// All techniques should have produced results
	if len(ctx.Directions) == 0 {
		t.Error("no primary directions")
	}
	if ctx.Progressions == nil {
		t.Error("no progressions")
	}
	if ctx.SolarReturn == nil {
		t.Error("no solar return")
	}
	if len(ctx.LunarReturns) == 0 {
		t.Error("no lunar returns")
	}
	if ctx.Profection == nil {
		t.Error("no profection")
	}
	if ctx.Firdaria == nil {
		t.Error("no firdaria")
	}
	if ctx.ZRFortune == nil {
		t.Error("no ZR Fortune")
	}

	// Brief should be non-empty and contain key sections
	if ctx.Brief == "" {
		t.Fatal("brief is empty")
	}
	if !strings.Contains(ctx.Brief, "SEÑORES DEL TIEMPO") {
		t.Error("brief missing time lords section")
	}
	if !strings.Contains(ctx.Brief, "DIRECCIONES PRIMARIAS") {
		t.Error("brief missing directions section")
	}
	if !strings.Contains(ctx.Brief, "CONVERGENCIA") {
		t.Error("brief missing convergence matrix")
	}

	t.Logf("Brief length: %d chars", len(ctx.Brief))
	t.Logf("Directions: %d, SA: %d, Eclipses: %d, Lunar Returns: %d",
		len(ctx.Directions), len(ctx.SolarArc), len(ctx.Eclipses), len(ctx.LunarReturns))

	// Print first 500 chars of brief
	preview := ctx.Brief
	if len(preview) > 500 {
		preview = preview[:500] + "..."
	}
	t.Logf("\n%s", preview)
}

func TestBuildBrief_Sections(t *testing.T) {
	chart := adrianChart(t)
	birthDate := time.Date(1975, 12, 27, 0, 0, 0, 0, time.UTC)

	ctx, err := Build(chart, "Adrian", birthDate, 2026)
	if err != nil {
		t.Fatal(err)
	}

	sections := []string{
		"SEÑORES DEL TIEMPO",
		"DIRECCIONES PRIMARIAS",
		"ARCOS SOLARES",
		"PROGRESIONES SECUNDARIAS",
		"ECLIPSES",
		"REVOLUCIÓN SOLAR",
		"CONVERGENCIA MENSUAL",
	}
	for _, s := range sections {
		if !strings.Contains(ctx.Brief, s) {
			t.Errorf("brief missing section: %s", s)
		}
	}
}
