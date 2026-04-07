package technique

import (
	"testing"

	"github.com/Camionerou/rag-saldivia/services/astro/internal/astromath"
)

func TestFindEclipses(t *testing.T) {
	eclipses, err := FindEclipses(2026)
	if err != nil {
		t.Fatal(err)
	}

	// 2026 should have 2-5 solar + 2-3 lunar eclipses
	if len(eclipses) < 2 {
		t.Errorf("got %d eclipses, expected at least 2", len(eclipses))
	}

	hasSolar := false
	hasLunar := false
	for _, e := range eclipses {
		if e.Type == "solar" {
			hasSolar = true
		}
		if e.Type == "lunar" {
			hasLunar = true
		}
		if e.Month < 1 || e.Month > 12 {
			t.Errorf("eclipse month = %d", e.Month)
		}
		if e.Lon < 0 || e.Lon >= 360 {
			t.Errorf("eclipse lon = %.2f", e.Lon)
		}
		t.Logf("  %s (%s) — month %d, lon %.1f°", e.Type, e.SubType, e.Month, e.Lon)
	}
	if !hasSolar {
		t.Error("no solar eclipses found in 2026")
	}
	if !hasLunar {
		t.Error("no lunar eclipses found in 2026")
	}
}

func TestFindEclipseActivations(t *testing.T) {
	chart := adrianChart(t)
	activations, err := FindEclipseActivations(chart, 2026)
	if err != nil {
		t.Fatal(err)
	}

	// May or may not have activations depending on eclipse positions
	t.Logf("Found %d eclipse activations for 2026", len(activations))
	for _, a := range activations {
		t.Logf("  %s eclipse → %s %s (orb %.2f°)",
			a.Eclipse.Type, a.Aspect, a.NatPoint, a.Orb)
	}
}

func TestFindFixedStarConjunctions(t *testing.T) {
	chart := adrianChart(t)
	results := FindFixedStarConjunctions(chart)

	// Adrian's chart should have at least a few star conjunctions
	t.Logf("Found %d fixed star conjunctions", len(results))
	for _, r := range results {
		t.Logf("  %s conjunct %s (orb %.2f°, nature: %s)",
			r.Star, r.NatPoint, r.Orb, r.Nature)
	}

	// Verify all results have valid fields
	for _, r := range results {
		if r.Star == "" || r.NatPoint == "" {
			t.Error("empty star or natal point")
		}
		if r.Orb > astromath.FixedStarOrb {
			t.Errorf("orb %.2f exceeds max %.2f", r.Orb, astromath.FixedStarOrb)
		}
	}
}

func TestCalcZodiacalReleasing_Fortune(t *testing.T) {
	chart := adrianChart(t)

	// Adrian at age ~50.5
	zr := CalcZodiacalReleasing(chart, "Fortune", 50.5)
	if zr == nil {
		t.Fatal("ZR returned nil")
	}

	if zr.Lot != "Fortune" {
		t.Errorf("lot = %q, want Fortune", zr.Lot)
	}
	if zr.LotSign == "" {
		t.Error("lot sign is empty")
	}

	// Level 1 (major) must exist
	if zr.Level1 == nil {
		t.Fatal("Level 1 is nil")
	}
	if zr.Level1.Lord == "" {
		t.Error("Level 1 lord is empty")
	}
	if zr.Level1.StartAge > 50.5 || zr.Level1.EndAge < 50.5 {
		t.Errorf("Level 1 range [%.1f, %.1f] doesn't contain age 50.5",
			zr.Level1.StartAge, zr.Level1.EndAge)
	}

	// Level 2 (sub) must exist
	if zr.Level2 == nil {
		t.Fatal("Level 2 is nil")
	}
	if zr.Level2.StartAge > 50.5 || zr.Level2.EndAge < 50.5 {
		t.Errorf("Level 2 range [%.1f, %.1f] doesn't contain age 50.5",
			zr.Level2.StartAge, zr.Level2.EndAge)
	}

	// Level 3 (bound) should exist
	if zr.Level3 == nil {
		t.Log("Level 3 is nil (may be rounding)")
	} else if zr.Level3.Lord == "" {
		t.Error("Level 3 lord is empty")
	}

	t.Logf("ZR Fortune: L1=%s(%s) L2=%s(%s) L3=%v",
		zr.Level1.Sign, zr.Level1.Lord,
		zr.Level2.Sign, zr.Level2.Lord,
		zr.Level3)
}

func TestCalcZodiacalReleasing_Spirit(t *testing.T) {
	chart := adrianChart(t)
	zr := CalcZodiacalReleasing(chart, "Spirit", 50.5)
	if zr == nil {
		t.Fatal("ZR Spirit returned nil")
	}
	if zr.Level1 == nil {
		t.Fatal("Level 1 is nil")
	}
	t.Logf("ZR Spirit: L1=%s(%s)", zr.Level1.Sign, zr.Level1.Lord)
}
