package ephemeris

import (
	"math"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	p := os.Getenv("EPHE_PATH")
	Init(p) // empty string = Moshier fallback
	code := m.Run()
	Close()
	os.Exit(code)
}

func TestJulDay_KnownDate(t *testing.T) {
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
	jd := JulDay(2000, 1, 1, 12.0)
	pos, err := CalcPlanet(jd, Sun, FlagSwieph|FlagSpeed)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(pos.Lon-280.46) > 0.1 {
		t.Errorf("Sun lon = %.4f, want ~280.46", pos.Lon)
	}
	if pos.Speed < 0.9 || pos.Speed > 1.1 {
		t.Errorf("Sun speed = %.4f, want ~1.0 deg/day", pos.Speed)
	}
}

func TestCalcPlanet_Equatorial(t *testing.T) {
	jd := JulDay(2000, 1, 1, 12.0)
	pos, err := CalcPlanet(jd, Sun, FlagSwieph|FlagSpeed|FlagEquatorial)
	if err != nil {
		t.Fatal(err)
	}
	if pos.RA < 270 || pos.RA > 290 {
		t.Errorf("Sun RA = %.4f, expected ~281", pos.RA)
	}
	// With equatorial flag, ecliptic fields should be zeroed
	if pos.Lon != 0 {
		t.Errorf("Lon should be 0 with equatorial flag, got %.4f", pos.Lon)
	}
}

func TestCalcPlanetFull(t *testing.T) {
	jd := JulDay(2000, 1, 1, 12.0)
	pos, err := CalcPlanetFull(jd, Sun, FlagSwieph|FlagSpeed)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(pos.Lon-280.46) > 0.1 {
		t.Errorf("Sun lon = %.4f, want ~280.46", pos.Lon)
	}
	if pos.RA < 270 || pos.RA > 290 {
		t.Errorf("Sun RA = %.4f, expected ~281", pos.RA)
	}
	if pos.Dec > -20 || pos.Dec < -25 {
		t.Errorf("Sun Dec = %.4f, expected ~-23", pos.Dec)
	}
}

func TestCalcHouses_TopocentricoRosario(t *testing.T) {
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
	jd := JulDay(2000, 1, 1, 12.0)
	eps, err := EclNut(jd)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(eps-23.44) > 0.1 {
		t.Errorf("epsilon = %.4f, want ~23.44", eps)
	}
}

func TestSolcrossUT(t *testing.T) {
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

func TestCstr(t *testing.T) {
	tests := []struct {
		in   []byte
		want string
	}{
		{[]byte("hello\x00\x00\x00"), "hello"},
		{[]byte{0, 0, 0}, ""},
		{[]byte("no null"), "no null"},
	}
	for _, tc := range tests {
		got := cstr(tc.in)
		if got != tc.want {
			t.Errorf("cstr(%v) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
