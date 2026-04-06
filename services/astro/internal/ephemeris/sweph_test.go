package ephemeris

import (
	"math"
	"os"
	"testing"
)

func ephePath(t *testing.T) string {
	t.Helper()
	p := os.Getenv("EPHE_PATH")
	if p == "" {
		// Use empty string — swephgo falls back to Moshier (sufficient for tests)
		return ""
	}
	return p
}

func TestJulDay_KnownDate(t *testing.T) {
	// J2000.0 epoch: 2000-01-01 12:00 UT = JD 2451545.0
	jd := JulDay(2000, 1, 1, 12.0)
	if math.Abs(jd-2451545.0) > 0.0001 {
		t.Errorf("JulDay = %.4f, want 2451545.0", jd)
	}
}

func TestRevJul_RoundTrip(t *testing.T) {
	jd := JulDay(1975, 12, 27, 19.233)
	y, m, d, h := RevJul(jd)
	if y != 1975 || m != 12 || d != 27 {
		t.Errorf("RevJul = %d-%d-%d, want 1975-12-27", y, m, d)
	}
	if math.Abs(h-19.233) > 0.01 {
		t.Errorf("RevJul hour = %.3f, want ~19.233", h)
	}
}

func TestCalcPlanet_SunPosition(t *testing.T) {
	Init(ephePath(t))
	defer Close()

	jd := JulDay(2000, 1, 1, 12.0)
	pos, err := CalcPlanet(jd, Sun, FlagSwieph|FlagSpeed)
	if err != nil {
		t.Fatal(err)
	}
	// Sun at J2000.0 ≈ 280.46° ecliptic longitude
	if math.Abs(pos.Lon-280.46) > 0.1 {
		t.Errorf("Sun lon = %.4f, want ~280.46", pos.Lon)
	}
	if pos.Speed < 0.9 || pos.Speed > 1.1 {
		t.Errorf("Sun speed = %.4f, want ~1.0 deg/day", pos.Speed)
	}
}

func TestCalcPlanetFull_Equatorial(t *testing.T) {
	Init(ephePath(t))
	defer Close()

	jd := JulDay(2000, 1, 1, 12.0)
	pos, err := CalcPlanetFull(jd, Sun, FlagSwieph|FlagSpeed)
	if err != nil {
		t.Fatal(err)
	}
	// Ecliptic lon populated
	if math.Abs(pos.Lon-280.46) > 0.1 {
		t.Errorf("Sun lon = %.4f, want ~280.46", pos.Lon)
	}
	// RA should also be populated (~281° for J2000.0 Sun)
	if pos.RA < 270 || pos.RA > 290 {
		t.Errorf("Sun RA = %.4f, expected ~281", pos.RA)
	}
}

func TestCalcHouses_TopocentricoRosario(t *testing.T) {
	Init(ephePath(t))
	defer Close()

	// Adrian's birth: 1975-12-27 19:14 UT, Rosario
	jd := JulDay(1975, 12, 27, 19.0+14.0/60.0)
	cusps, ascmc, err := CalcHouses(jd, -32.9468, -60.6393, HouseTopocentric)
	if err != nil {
		t.Fatal(err)
	}
	asc := ascmc[0]
	if asc < 30 || asc > 100 {
		t.Errorf("ASC = %.2f, expected ~45-75 range", asc)
	}
	if len(cusps) < 13 {
		t.Errorf("got %d cusps, want 13", len(cusps))
	}
}

func TestEclNut(t *testing.T) {
	Init(ephePath(t))
	defer Close()

	jd := JulDay(2000, 1, 1, 12.0)
	eps, err := EclNut(jd)
	if err != nil {
		t.Fatal(err)
	}
	// Obliquity ~23.44° at J2000.0
	if math.Abs(eps-23.44) > 0.1 {
		t.Errorf("epsilon = %.4f, want ~23.44", eps)
	}
}

func TestSolcrossUT(t *testing.T) {
	Init(ephePath(t))
	defer Close()

	// Find when Sun crosses 0° (vernal equinox) in 2026
	jdStart := JulDay(2026, 3, 1, 0.0)
	jd, err := SolcrossUT(0.0, jdStart, FlagSwieph|FlagSpeed)
	if err != nil {
		t.Fatal(err)
	}
	y, m, d, _ := RevJul(jd)
	if y != 2026 || m != 3 || (d < 19 || d > 21) {
		t.Errorf("vernal equinox 2026 = %d-%d-%d, want ~2026-03-20", y, m, d)
	}
}
