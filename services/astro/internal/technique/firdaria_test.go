package technique

import (
	"encoding/json"
	"math"
	"os"
	"testing"
	"time"
)

func TestCalcFirdaria_Golden(t *testing.T) {
	data, err := os.ReadFile("../../testdata/golden/firdaria_adrian_2026.json")
	if err != nil {
		t.Skip("golden file not found")
	}

	var golden struct {
		Output struct {
			Diurnal       bool    `json:"diurnal"`
			Age           float64 `json:"age"`
			MajorLord     string  `json:"major_lord"`
			MajorYears    int     `json:"major_years"`
			MajorStartAge float64 `json:"major_start_age"`
			MajorEndAge   float64 `json:"major_end_age"`
			SubLord       string  `json:"sub_lord"`
			NextMajorLord string  `json:"next_major_lord"`
		} `json:"output"`
	}
	if err := json.Unmarshal(data, &golden); err != nil {
		t.Fatal(err)
	}

	birthDate := time.Date(1975, 12, 27, 0, 0, 0, 0, time.UTC)
	fird := CalcFirdaria(birthDate, golden.Output.Diurnal, 2026)

	if fird.MajorLord != golden.Output.MajorLord {
		t.Errorf("major_lord = %q, want %q", fird.MajorLord, golden.Output.MajorLord)
	}
	if fird.MajorYears != golden.Output.MajorYears {
		t.Errorf("major_years = %d, want %d", fird.MajorYears, golden.Output.MajorYears)
	}
	if fird.SubLord != golden.Output.SubLord {
		t.Errorf("sub_lord = %q, want %q", fird.SubLord, golden.Output.SubLord)
	}
	if math.Abs(fird.MajorStartAge-golden.Output.MajorStartAge) > 0.5 {
		t.Errorf("major_start_age = %.1f, want %.1f", fird.MajorStartAge, golden.Output.MajorStartAge)
	}
	if fird.NextMajorLord != golden.Output.NextMajorLord {
		t.Errorf("next_major_lord = %q, want %q", fird.NextMajorLord, golden.Output.NextMajorLord)
	}
}

func TestCalcFirdaria_DiurnalSequence(t *testing.T) {
	birthDate := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

	// Age 0 → Sol (first diurnal period)
	fird := CalcFirdaria(birthDate, true, 2000)
	if fird.MajorLord != "Sol" {
		t.Errorf("age ~0 major = %q, want Sol", fird.MajorLord)
	}

	// Age 11 → Venus (second period, Sol=10y)
	fird = CalcFirdaria(birthDate, true, 2011)
	if fird.MajorLord != "Venus" {
		t.Errorf("age ~11 major = %q, want Venus", fird.MajorLord)
	}
}
