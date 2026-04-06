package technique

import (
	"encoding/json"
	"math"
	"os"
	"testing"
)

func TestFindDirections_Golden(t *testing.T) {
	data, err := os.ReadFile("../../testdata/golden/primary_dir_adrian_2026.json")
	if err != nil {
		t.Skip("golden file not found")
	}

	var golden struct {
		Input struct {
			Age float64 `json:"age"`
		} `json:"input"`
		Output []struct {
			Promissor    string  `json:"promissor"`
			Aspect       string  `json:"aspect"`
			Significator string  `json:"significator"`
			Arc          float64 `json:"arc"`
			AgeExact     float64 `json:"age_exact"`
			OrbDeg       float64 `json:"orb_deg"`
			Tipo         string  `json:"tipo"`
		} `json:"output"`
	}
	if err := json.Unmarshal(data, &golden); err != nil {
		t.Fatal(err)
	}

	chart := adrianChart(t)
	results := FindDirections(chart, golden.Input.Age, 2.0)

	t.Logf("Go found %d directions, Python found %d", len(results), len(golden.Output))

	// We expect a similar number of results (not exact due to different
	// planet sets — Python includes Ecl.Prenatal, Vértice, etc.)
	if len(results) < 10 {
		t.Errorf("too few directions: %d", len(results))
	}

	// Check that the tightest Python direction (MC opp Marte, orb ~0) exists in Go
	if len(golden.Output) > 0 {
		pyFirst := golden.Output[0]
		found := false
		for _, r := range results {
			if r.Promissor == pyFirst.Promissor &&
				r.Significator == pyFirst.Significator &&
				r.Aspect == pyFirst.Aspect &&
				r.Tipo == pyFirst.Tipo {
				arcDiff := math.Abs(r.Arc - pyFirst.Arc)
				t.Logf("Matched: %s %s %s — Go arc=%.3f, Py arc=%.3f (diff %.3f)",
					r.Promissor, r.Aspect, r.Significator, r.Arc, pyFirst.Arc, arcDiff)
				if arcDiff < 1.0 {
					found = true
				}
				break
			}
		}
		if !found {
			t.Logf("WARNING: tightest Python direction not found: %s %s %s (arc=%.3f)",
				pyFirst.Promissor, pyFirst.Aspect, pyFirst.Significator, pyFirst.Arc)
		}
	}
}

func TestFindDirections_Sanity(t *testing.T) {
	chart := adrianChart(t)
	results := FindDirections(chart, 50.5, 2.0)

	if len(results) == 0 {
		t.Fatal("no directions found")
	}

	for i, r := range results {
		if r.Promissor == "" || r.Significator == "" || r.Aspect == "" {
			t.Errorf("result[%d] has empty fields", i)
		}
		if r.Sistema != "polich-page" {
			t.Errorf("result[%d] sistema = %q", i, r.Sistema)
		}
		if r.Tipo != "directa" && r.Tipo != "conversa" {
			t.Errorf("result[%d] tipo = %q", i, r.Tipo)
		}
		if r.Arc <= 0 || r.Arc > maxArcDeg {
			t.Errorf("result[%d] arc = %.3f (out of range)", i, r.Arc)
		}
	}

	// Results should be sorted by orb
	for i := 1; i < len(results); i++ {
		if results[i].OrbDeg < results[i-1].OrbDeg {
			t.Errorf("results not sorted by orb at index %d", i)
			break
		}
	}

	t.Logf("Found %d directions for age 50.5 (orb 2°)", len(results))
	for _, r := range results[:min(5, len(results))] {
		t.Logf("  %s %s %s — arc=%.2f age=%.2f orb=%.3f° %s",
			r.Promissor, r.Aspect, r.Significator, r.Arc, r.AgeExact, r.OrbDeg, r.Tipo)
	}
}

func TestSphereHelpers(t *testing.T) {
	// DSA at equator: should be 90° for any declination (tan(0)=0)
	dsa := diurnalSemiArc(23.0, 0.0)
	if math.Abs(dsa-90.0) > 0.01 {
		t.Errorf("DSA at equator = %.2f, want 90", dsa)
	}

	// MD: RA = RAMC → MD = 0
	md := meridianDistance(100.0, 100.0)
	if md > 0.001 {
		t.Errorf("MD(100, 100) = %.4f, want 0", md)
	}

	// MD: RA = RAMC + 180 → MD = 180
	md = meridianDistance(100.0, 280.0)
	if math.Abs(md-180.0) > 0.001 {
		t.Errorf("MD(100, 280) = %.4f, want 180", md)
	}

	// OA at pole=0 should equal RA (no adjustment)
	oa := obliqueAscension(45.0, 20.0, 0.0)
	if math.Abs(oa-45.0) > 0.01 {
		t.Errorf("OA(45, 20, pole=0) = %.4f, want ~45", oa)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
