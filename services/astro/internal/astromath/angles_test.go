package astromath

import (
	"math"
	"testing"
)

func TestAngDiff(t *testing.T) {
	tests := []struct {
		a, b, want float64
	}{
		{350, 10, 20}, {10, 350, 20}, {90, 270, 180},
		{0, 180, 180}, {0, 0, 0}, {100, 100.5, 0.5},
	}
	for _, tc := range tests {
		got := AngDiff(tc.a, tc.b)
		if math.Abs(got-tc.want) > 0.001 {
			t.Errorf("AngDiff(%.1f, %.1f) = %.3f, want %.3f", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestNormalize360(t *testing.T) {
	tests := []struct{ in, want float64 }{
		{370, 10}, {-10, 350}, {720, 0}, {0, 0}, {359.99, 359.99},
	}
	for _, tc := range tests {
		got := Normalize360(tc.in)
		if math.Abs(got-tc.want) > 0.001 {
			t.Errorf("Normalize360(%.2f) = %.3f, want %.3f", tc.in, got, tc.want)
		}
	}
}

func TestPosToStr(t *testing.T) {
	tests := []struct {
		lon  float64
		want string
	}{
		{0, "0°00' Aries"},
		{30, "0°00' Tauro"},
		{95.5, "5°30' Cáncer"},
		{275.75, "5°45' Capricornio"},
	}
	for _, tc := range tests {
		got := PosToStr(tc.lon)
		if got != tc.want {
			t.Errorf("PosToStr(%.2f) = %q, want %q", tc.lon, got, tc.want)
		}
	}
}

func TestSignIndex(t *testing.T) {
	if SignIndex(0) != 0 {
		t.Error("0° should be Aries (0)")
	}
	if SignIndex(30) != 1 {
		t.Error("30° should be Taurus (1)")
	}
	if SignIndex(359) != 11 {
		t.Error("359° should be Pisces (11)")
	}
}

func TestEclToEq(t *testing.T) {
	ra, dec := EclToEq(0, 0, 23.4393)
	if math.Abs(ra) > 0.01 || math.Abs(dec) > 0.01 {
		t.Errorf("EclToEq(0, 0) = (%.4f, %.4f), want (0, 0)", ra, dec)
	}
}

func TestPartOfFortune(t *testing.T) {
	pof := PartOfFortune(90, 120, 30, true)
	want := Normalize360(90 + 120 - 30)
	if math.Abs(pof-want) > 0.001 {
		t.Errorf("PartOfFortune diurnal = %.2f, want %.2f", pof, want)
	}
}

func TestCombustionStatus(t *testing.T) {
	if CombustionStatus(100, 100.1) != "cazimi" {
		t.Error("0.1° from Sun should be cazimi")
	}
	if CombustionStatus(100, 105) != "combust" {
		t.Error("5° from Sun should be combust")
	}
	if CombustionStatus(100, 120) != "" {
		t.Error("20° from Sun should not be combust")
	}
}

func TestBoundLord(t *testing.T) {
	// 0° Aries → Jupiter (first bound)
	if got := BoundLord(0); got != "Júpiter" {
		t.Errorf("BoundLord(0) = %q, want Júpiter", got)
	}
	// 7° Aries → Venus (second bound: 6-12)
	if got := BoundLord(7); got != "Venus" {
		t.Errorf("BoundLord(7) = %q, want Venus", got)
	}
}

func TestFindAspect(t *testing.T) {
	asp := FindAspect(0, 90, 3.0)
	if asp == nil || asp.Name != "square" {
		t.Error("0° and 90° should form a square")
	}
	asp = FindAspect(0, 45, 3.0)
	if asp != nil {
		t.Error("0° and 45° should not form a major aspect")
	}
}

func TestAntiscion(t *testing.T) {
	// 0° Aries → 180° Virgo (antiscion)
	got := Antiscion(0)
	if math.Abs(got-180) > 0.001 {
		t.Errorf("Antiscion(0) = %.2f, want 180", got)
	}
}
