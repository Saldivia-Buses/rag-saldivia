package astromath

import "testing"

func TestWorldCities_Count(t *testing.T) {
	if got := len(WorldCities); got < 900 {
		t.Errorf("WorldCities count = %d, want >= 900", got)
	}
	t.Logf("total cities: %d", len(WorldCities))
}

func TestLookupCity_ExactMatch(t *testing.T) {
	tests := []struct {
		query string
		want  string
	}{
		{"buenos aires", "Buenos Aires"},
		{"Madrid", "Madrid"},
		{"new york", "New York"},
		{"santiago", "Santiago"},
		{"tokyo", "Tokyo"},
		{"london", "London"},
		{"paris", "Paris"},
		{"sydney", "Sydney"},
	}

	for _, tt := range tests {
		loc := LookupCity(tt.query)
		if loc == nil {
			t.Errorf("LookupCity(%q) = nil, want %q", tt.query, tt.want)
			continue
		}
		if loc.City != tt.want {
			t.Errorf("LookupCity(%q).City = %q, want %q", tt.query, loc.City, tt.want)
		}
	}
}

func TestLookupCity_CaseInsensitive(t *testing.T) {
	loc := LookupCity("BUENOS AIRES")
	if loc == nil {
		t.Fatal("LookupCity(BUENOS AIRES) = nil")
	}
	if loc.City != "Buenos Aires" {
		t.Errorf("City = %q, want Buenos Aires", loc.City)
	}
}

func TestLookupCity_AccentNormalization(t *testing.T) {
	// "córdoba" with accent should find "cordoba" in the map
	loc := LookupCity("córdoba")
	if loc == nil {
		t.Fatal("LookupCity(córdoba) = nil")
	}
}

func TestLookupCity_NotFound(t *testing.T) {
	loc := LookupCity("narnia")
	if loc != nil {
		t.Errorf("LookupCity(narnia) = %v, want nil", loc)
	}
}

func TestWorldCities_UTCOffsets(t *testing.T) {
	tests := []struct {
		key  string
		want float64
	}{
		{"buenos aires", -3},
		{"new york", -5},
		{"london", 0},
		{"tokyo", 9},
		{"mumbai", 5.5},    // half-hour timezone
		{"kathmandu", 5.75}, // quarter-hour timezone
		{"tehran", 3.5},     // half-hour timezone
	}

	for _, tt := range tests {
		loc, ok := WorldCities[tt.key]
		if !ok {
			t.Errorf("WorldCities[%q] not found", tt.key)
			continue
		}
		if loc.UTCOff != tt.want {
			t.Errorf("WorldCities[%q].UTCOff = %v, want %v", tt.key, loc.UTCOff, tt.want)
		}
	}
}

func TestWorldCities_Coordinates(t *testing.T) {
	// Spot check a few cities for reasonable coordinates
	tests := []struct {
		key    string
		latMin float64
		latMax float64
	}{
		{"buenos aires", -35, -34},
		{"london", 51, 52},
		{"tokyo", 35, 36},
		{"sydney", -34, -33},
	}

	for _, tt := range tests {
		loc, ok := WorldCities[tt.key]
		if !ok {
			t.Errorf("WorldCities[%q] not found", tt.key)
			continue
		}
		if loc.Lat < tt.latMin || loc.Lat > tt.latMax {
			t.Errorf("WorldCities[%q].Lat = %v, want between %v and %v", tt.key, loc.Lat, tt.latMin, tt.latMax)
		}
	}
}
