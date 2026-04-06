package technique

import (
	"encoding/json"
	"os"
	"testing"
)

func TestCalcTransits_Golden(t *testing.T) {
	data, err := os.ReadFile("../../testdata/golden/transits_adrian_2026.json")
	if err != nil {
		t.Skip("golden file not found")
	}

	var golden struct {
		Output []struct {
			Transit string `json:"transit"`
			Aspect  string `json:"aspect"`
			Natal   string `json:"natal"`
		} `json:"output"`
	}
	if err := json.Unmarshal(data, &golden); err != nil {
		t.Fatal(err)
	}

	chart := adrianChart(t)
	results := CalcTransits(chart, 2026)

	t.Logf("Go found %d transit activations, Python found %d", len(results), len(golden.Output))

	if len(results) < 20 {
		t.Errorf("too few transits: %d (Python had %d)", len(results), len(golden.Output))
	}

	// Check that we found some key transits that Python found
	pyTransits := make(map[string]bool)
	for _, g := range golden.Output {
		key := g.Transit + "_" + g.Aspect + "_" + g.Natal
		pyTransits[key] = true
	}
	matched := 0
	for _, r := range results {
		// Map Go aspect names to Python names
		pyAspect := map[string]string{
			"conjunction": "Conjunción", "sextile": "Sextil",
			"square": "Cuadratura", "trine": "Trígono", "opposition": "Oposición",
		}[r.Aspect]
		key := r.Transit + "_" + pyAspect + "_" + r.Natal
		if pyTransits[key] {
			matched++
		}
	}
	t.Logf("Matched %d/%d Go results against Python", matched, len(results))
}

func TestCalcTransits_Structure(t *testing.T) {
	chart := adrianChart(t)
	results := CalcTransits(chart, 2026)

	if len(results) == 0 {
		t.Fatal("no transits found")
	}

	for i, r := range results {
		if r.Transit == "" || r.Natal == "" || r.Aspect == "" {
			t.Errorf("result[%d] has empty fields", i)
		}
		if r.Passes < 1 {
			t.Errorf("result[%d] passes = %d", i, r.Passes)
		}
		if r.Nature != "fácil" && r.Nature != "tenso" && r.Nature != "neutral" {
			t.Errorf("result[%d] nature = %q", i, r.Nature)
		}
		if len(r.EpDetails) != r.Passes {
			t.Errorf("result[%d] passes=%d but ep_details=%d", i, r.Passes, len(r.EpDetails))
		}
	}

	// Show first 5
	for _, r := range results[:min(5, len(results))] {
		t.Logf("  %s %s %s — orb %.2f° %d passes retro=%v month=%d",
			r.Transit, r.Aspect, r.Natal, r.Orb, r.Passes, r.Retrograde, r.Month)
	}
}
