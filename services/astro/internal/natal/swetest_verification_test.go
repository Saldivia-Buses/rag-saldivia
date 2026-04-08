package natal

import (
	"math"
	"os"
	"testing"

	"github.com/Camionerou/rag-saldivia/services/astro/internal/ephemeris"
)

// TestBuildNatal_SwetestVerification compares our BuildNatal output against
// swetest CLI (Swiss Ephemeris reference implementation) using Moshier ephemeris.
//
// Birth data: Adrian Saldivia
//   Date: 27/12/1975, 16:14 local (UTC-3) = 19:14 UT
//   Place: Rosario, Argentina (-32.9468, -60.6393, 25m)
//   House system: Topocentric (Polich-Page, 'T')
//
// swetest command:
//   swetest -b27.12.1975 -ut19:14 -geopos-60.6393,-32.9468,25 -p0123456789 -house-60.6393,-32.9468,T
//
// Reference values from swetest with Moshier ephemeris (no .se1 files).
func TestBuildNatal_SwetestVerification(t *testing.T) {
	ephePath := os.Getenv("EPHE_PATH")
	if ephePath == "" {
		ephePath = "/ephe"
	}
	ephemeris.Init(ephePath)
	defer ephemeris.Close()

	// Adrian Saldivia: 27/12/1975, 16:14 local, Rosario (-32.9468, -60.6393, 25m), UTC-3
	chart, err := BuildNatal(1975, 12, 27, 16.0+14.0/60.0, -32.9468, -60.6393, 25.0, -3)
	if err != nil {
		t.Fatalf("BuildNatal failed: %v", err)
	}

	// swetest reference positions (Moshier ephemeris, TOPOCENTRIC)
	// Command: swetest -b27.12.1975 -ut19:14 -topo-60.6393,-32.9468,25 -p0123456789t -fPl
	// IMPORTANT: these are TOPOCENTRIC positions (observer-corrected), not geocentric.
	// The Moon difference between geo and topo is ~0.95° (parallax).
	swetest := map[string]float64{
		"Sol":        275.4094460,
		"Luna":       213.0501069,
		"Mercurio":   291.2691684,
		"Venus":      234.3631784,
		"Marte":       78.5872595,
		"Júpiter":     15.2600721,
		"Saturno":    121.3600977,
		"Urano":      216.2232475,
		"Neptuno":    252.3803154,
		"Plutón":     191.6184650,
		"Nodo Norte": 231.0008237, // true node (not affected by topo)
	}

	// Swetest house cusps (Topocentric/Polich-Page)
	swetestCusps := map[int]float64{
		1:  45.0184902,  // 45°01'06.56"
		2:  74.4974922,  // 74°29'50.97"
		3:  106.9816844, // 106°58'54.06"
		4:  141.1220654, // 141°07'19.44"
		5:  173.5341802, // 173°32'03.05"
		6:  201.4893130, // 201°29'21.53"
		7:  225.0184902,
		8:  254.4974922,
		9:  286.9816844,
		10: 321.1220654, // MC
		11: 353.5341802,
		12: 21.4893130,
	}

	swetestASC := 45.0184902
	swetestMC := 321.1220654

	// Tolerance: 0.01° (36 arcseconds) — accounts for topocentric correction differences
	// between our wrapper and swetest direct call. For geocentric positions, 0.001° is achievable.
	tolerance := 0.01

	// Verify planet positions
	for name, expected := range swetest {
		pos, ok := chart.Planets[name]
		if !ok {
			t.Errorf("Planet %s not found in chart", name)
			continue
		}
		diff := math.Abs(pos.Lon - expected)
		if diff > 180 { diff = 360 - diff }
		if diff > tolerance {
			t.Errorf("Planet %s: got %.6f°, swetest %.6f°, diff %.6f° (tolerance %.3f°)",
				name, pos.Lon, expected, diff, tolerance)
		} else {
			t.Logf("✓ %s: %.6f° (swetest: %.6f°, diff: %.6f°)", name, pos.Lon, expected, diff)
		}
	}

	// Verify ASC
	ascDiff := math.Abs(chart.ASC - swetestASC)
	if ascDiff > 180 { ascDiff = 360 - ascDiff }
	if ascDiff > tolerance {
		t.Errorf("ASC: got %.6f°, swetest %.6f°, diff %.6f°", chart.ASC, swetestASC, ascDiff)
	} else {
		t.Logf("✓ ASC: %.6f° (swetest: %.6f°, diff: %.6f°)", chart.ASC, swetestASC, ascDiff)
	}

	// Verify MC
	mcDiff := math.Abs(chart.MC - swetestMC)
	if mcDiff > 180 { mcDiff = 360 - mcDiff }
	if mcDiff > tolerance {
		t.Errorf("MC: got %.6f°, swetest %.6f°, diff %.6f°", chart.MC, swetestMC, mcDiff)
	} else {
		t.Logf("✓ MC: %.6f° (swetest: %.6f°, diff: %.6f°)", chart.MC, swetestMC, mcDiff)
	}

	// Verify house cusps
	for house, expected := range swetestCusps {
		if house < 1 || house > 12 || len(chart.Cusps) <= house {
			continue
		}
		got := chart.Cusps[house]
		diff := math.Abs(got - expected)
		if diff > 180 { diff = 360 - diff }
		if diff > tolerance {
			t.Errorf("House %d: got %.6f°, swetest %.6f°, diff %.6f°", house, got, expected, diff)
		}
	}
}
