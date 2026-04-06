package technique

import (
	"encoding/json"
	"os"
	"testing"
	"time"
)

func TestCalcProfection_Golden(t *testing.T) {
	data, err := os.ReadFile("../../testdata/golden/profections_adrian_2026.json")
	if err != nil {
		t.Skip("golden file not found")
	}

	var golden struct {
		Output struct {
			Profection struct {
				Age        int    `json:"age"`
				CasaActiva int    `json:"casa_activa"`
				ProfSign   string `json:"prof_sign"`
				Lord       string `json:"lord"`
				Theme      string `json:"tema"`
			} `json:"profection"`
			Cascade struct {
				YearLord  string `json:"year_lord"`
				YearHouse int    `json:"year_house"`
			} `json:"cascade"`
		} `json:"output"`
	}
	if err := json.Unmarshal(data, &golden); err != nil {
		t.Fatal(err)
	}

	chart := adrianChart(t)
	birthDate := time.Date(1975, 12, 27, 0, 0, 0, 0, time.UTC)

	prof := CalcProfection(chart, birthDate, 2026)

	if prof.Age != golden.Output.Profection.Age {
		t.Errorf("age = %d, want %d", prof.Age, golden.Output.Profection.Age)
	}
	if prof.ActiveHouse != golden.Output.Profection.CasaActiva {
		t.Errorf("active_house = %d, want %d", prof.ActiveHouse, golden.Output.Profection.CasaActiva)
	}
	if prof.ProfSign != golden.Output.Profection.ProfSign {
		t.Errorf("prof_sign = %q, want %q", prof.ProfSign, golden.Output.Profection.ProfSign)
	}
	if prof.Lord != golden.Output.Profection.Lord {
		t.Errorf("lord = %q, want %q", prof.Lord, golden.Output.Profection.Lord)
	}

	// Cascade
	cascade := CalcProfectionCascade(chart, birthDate, 2026)
	if cascade.YearHouse != golden.Output.Cascade.YearHouse {
		t.Errorf("cascade year_house = %d, want %d", cascade.YearHouse, golden.Output.Cascade.YearHouse)
	}
}

func TestCalcProfection_AgeZero(t *testing.T) {
	chart := adrianChart(t)
	birthDate := time.Date(1975, 12, 27, 0, 0, 0, 0, time.UTC)

	prof := CalcProfection(chart, birthDate, 1976)
	if prof.Age != 0 {
		t.Errorf("age at birth year = %d, want 0", prof.Age)
	}
	if prof.ActiveHouse != 1 {
		t.Errorf("active house at age 0 = %d, want 1", prof.ActiveHouse)
	}
}
