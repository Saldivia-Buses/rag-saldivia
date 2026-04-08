package astromath

import (
	"testing"

	"github.com/Camionerou/rag-saldivia/services/astro/internal/ephemeris"
)

func TestCalcDisposition(t *testing.T) {
	// Sol in Leo (domicile → self-ruling)
	// Luna in Tauro (Venus rules Tauro → Venus in Libra → Venus self-ruling)
	planets := map[string]*ephemeris.PlanetPos{
		"Sol":      {Lon: 130}, // Leo 10° → ruler: Sol (self)
		"Luna":     {Lon: 45},  // Tauro 15° → ruler: Venus
		"Venus":    {Lon: 195}, // Libra 15° → ruler: Venus (self)
		"Marte":    {Lon: 280}, // Capricornio 10° → ruler: Saturno
		"Saturno":  {Lon: 310}, // Acuario 10° → ruler: Saturno (self)
		"Júpiter":  {Lon: 260}, // Sagitario 20° → ruler: Júpiter (self)
		"Mercurio": {Lon: 170}, // Virgo 20° → ruler: Mercurio (self)
	}

	result := CalcDisposition(planets)
	if result == nil {
		t.Fatal("CalcDisposition returned nil")
	}

	// Sol should have a complete chain ending at itself
	for _, c := range result.Chains {
		if c.Planet == "Sol" {
			if !c.Complete {
				t.Error("Sol in Leo should have complete chain (self-ruling)")
			}
			if c.Final != "Sol" {
				t.Errorf("Sol chain final = %q, want Sol", c.Final)
			}
		}
	}

	// Check chain count
	if len(result.Chains) == 0 {
		t.Error("no chains generated")
	}
}

func TestCalcDisposition_MutualReception(t *testing.T) {
	// Sol in Cáncer (ruler: Luna), Luna in Leo (ruler: Sol) → mutual reception
	planets := map[string]*ephemeris.PlanetPos{
		"Sol":  {Lon: 100}, // Cáncer 10°
		"Luna": {Lon: 130}, // Leo 10°
	}

	result := CalcDisposition(planets)
	if len(result.MutualReceptions) == 0 {
		t.Error("expected mutual reception between Sol and Luna")
	}
	if len(result.MutualReceptions) > 0 {
		mr := result.MutualReceptions[0]
		if (mr.PlanetA != "Sol" || mr.PlanetB != "Luna") && (mr.PlanetA != "Luna" || mr.PlanetB != "Sol") {
			t.Errorf("mutual reception: got %s-%s, want Sol-Luna", mr.PlanetA, mr.PlanetB)
		}
	}
}
