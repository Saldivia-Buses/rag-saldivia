package astromath

import "testing"

func TestSabianSymbol(t *testing.T) {
	tests := []struct {
		lon  float64
		want string
	}{
		{0, "A woman rises out of water, a seal rises and embraces her"},   // Aries 1°
		{29, "A duck pond and its brood"},                                   // Aries 30°
		{30, "A clear mountain stream"},                                     // Taurus 1°
		{359, "A majestic rock formation resembling a face is idealized by a boy who takes it as his ideal of greatness, and as he grows up, begins to look like it"}, // Pisces 30°
	}
	for _, tc := range tests {
		got := SabianSymbol(tc.lon)
		if got != tc.want {
			t.Errorf("SabianSymbol(%.0f) = %q, want %q", tc.lon, got, tc.want)
		}
	}
}

func TestSabianSymbol_AllPopulated(t *testing.T) {
	for i := 0; i < 360; i++ {
		s := SabianSymbol(float64(i))
		if s == "" {
			t.Errorf("SabianSymbol(%d) is empty", i)
		}
	}
}

func TestSabianDegree(t *testing.T) {
	tests := []struct {
		lon  float64
		want string
	}{
		{0, "Aries 1°"},
		{45, "Tauro 16°"},
		{359, "Piscis 30°"},
	}
	for _, tc := range tests {
		got := SabianDegree(tc.lon)
		if got != tc.want {
			t.Errorf("SabianDegree(%.0f) = %q, want %q", tc.lon, got, tc.want)
		}
	}
}
