package technique

import (
	"testing"

	"github.com/Camionerou/rag-saldivia/services/astro/internal/astromath"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/ephemeris"
)

// TestAllFixedStarsResolvable verifies every star in MajorFixedStars can be
// resolved by Swiss Ephemeris. This is a hard requirement (Plan 13, L2):
// if a SweName is wrong, the star silently disappears from conjunctions.
// Requires sefstars.txt in EPHE_PATH — skips in environments without it.
func TestAllFixedStarsResolvable(t *testing.T) {
	jd := ephemeris.JulDay(2026, 1, 1, 0.0)

	// Probe one known star to check if sefstars.txt is available
	_, err := ephemeris.FixstarUT("Regulus", jd, ephemeris.FlagSwieph)
	if err != nil {
		t.Skipf("sefstars.txt not available (need EPHE_PATH with fixed star catalog): %v", err)
	}

	for _, star := range astromath.MajorFixedStars {
		lon, err := ephemeris.FixstarUT(star.SweName, jd, ephemeris.FlagSwieph)
		if err != nil {
			t.Errorf("star %q (SweName=%q) not found in Swiss Ephemeris: %v", star.Name, star.SweName, err)
			continue
		}
		if lon < 0 || lon >= 360 {
			t.Errorf("star %q longitude out of range: %.4f", star.Name, lon)
		}
	}
}
