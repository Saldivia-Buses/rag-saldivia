package astromath

import "testing"

func TestMajorFixedStars_Count(t *testing.T) {
	if got := len(MajorFixedStars); got != 60 {
		t.Errorf("MajorFixedStars count = %d, want 60", got)
	}
}

func TestMajorFixedStars_UniqueSweName(t *testing.T) {
	seen := make(map[string]bool, len(MajorFixedStars))
	for _, s := range MajorFixedStars {
		if seen[s.SweName] {
			t.Errorf("duplicate SweName: %s", s.SweName)
		}
		seen[s.SweName] = true
	}
}

func TestMajorFixedStars_UniqueDisplayName(t *testing.T) {
	seen := make(map[string]bool, len(MajorFixedStars))
	for _, s := range MajorFixedStars {
		if seen[s.Name] {
			t.Errorf("duplicate display Name: %s", s.Name)
		}
		seen[s.Name] = true
	}
}

func TestMajorFixedStars_NatureNotEmpty(t *testing.T) {
	for _, s := range MajorFixedStars {
		if s.Nature == "" {
			t.Errorf("star %s has empty Nature", s.Name)
		}
	}
}
