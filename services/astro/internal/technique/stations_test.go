package technique

import (
	"testing"

	"github.com/Camionerou/rag-saldivia/services/astro/internal/astromath"
)

func TestFindStations(t *testing.T) {
	chart := adrianChart(t)
	stations := FindStations(chart, 2026)

	if len(stations) == 0 {
		t.Fatal("no stations found in 2026")
	}

	hasSR := false
	hasSD := false
	for _, st := range stations {
		if st.Type == "SR" {
			hasSR = true
		}
		if st.Type == "SD" {
			hasSD = true
		}
		_ = st.NatPoint
		if st.Month < 1 || st.Month > 12 {
			t.Errorf("station month = %d", st.Month)
		}
		t.Logf("  %s %s at %s (month %d day %d) natal=%s orb=%.1f°",
			st.Planet, st.Type, astromath.PosToStr(st.Lon), st.Month, st.Day, st.NatPoint, st.NatOrb)
	}

	if !hasSR {
		t.Error("no station retrograde found")
	}
	if !hasSD {
		t.Error("no station direct found")
	}

	t.Logf("Found %d stations (%d near natal points)", len(stations),
		func() int {
			n := 0
			for _, s := range stations {
				if s.NatPoint != "" {
					n++
				}
			}
			return n
		}())
}
