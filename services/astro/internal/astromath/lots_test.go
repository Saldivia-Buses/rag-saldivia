package astromath

import (
	"math"
	"testing"

	"github.com/Camionerou/rag-saldivia/services/astro/internal/ephemeris"
)

func TestCalcAllLots(t *testing.T) {
	// Synthetic chart: ASC=10°, Sol=120° (Leo), Luna=45° (Tauro), diurnal
	planets := map[string]*ephemeris.PlanetPos{
		"Sol":      {Lon: 120},
		"Luna":     {Lon: 45},
		"Venus":    {Lon: 200},
		"Marte":    {Lon: 300},
		"Júpiter":  {Lon: 150},
		"Saturno":  {Lon: 270},
		"Mercurio": {Lon: 100},
	}
	asc := 10.0
	cusps := make([]float64, 13)
	for i := range cusps {
		cusps[i] = Normalize360(float64(i) * 30)
	}

	lots := CalcAllLots(planets, asc, true, cusps)

	if len(lots) == 0 {
		t.Fatal("CalcAllLots returned no lots")
	}
	if len(lots) != len(LotDefinitions) {
		t.Errorf("CalcAllLots returned %d lots, want %d", len(lots), len(LotDefinitions))
	}

	// Check Fortune: diurnal = ASC + Luna - Sol = 10 + 45 - 120 = -65 → 295
	var fortuneLot *LotResult
	for i := range lots {
		if lots[i].Key == "fortune" {
			fortuneLot = &lots[i]
			break
		}
	}
	if fortuneLot == nil {
		t.Fatal("Fortune lot not found")
	}
	wantFortune := Normalize360(10 + 45 - 120) // 295
	if math.Abs(fortuneLot.Lon-wantFortune) > 0.01 {
		t.Errorf("Fortune = %.2f, want %.2f", fortuneLot.Lon, wantFortune)
	}

	// All lots should have valid sign names
	for _, lot := range lots {
		if lot.Sign == "" {
			t.Errorf("Lot %q has empty sign", lot.Key)
		}
		if lot.Lon < 0 || lot.Lon >= 360 {
			t.Errorf("Lot %q has invalid lon %.2f", lot.Key, lot.Lon)
		}
	}
}

func TestCalcLot_NocturnalReverse(t *testing.T) {
	planets := map[string]*ephemeris.PlanetPos{
		"Sol":  {Lon: 120},
		"Luna": {Lon: 45},
	}
	cusps := make([]float64, 13)
	asc := 10.0

	// Fortune nocturnal = ASC + Sol - Luna = 10 + 120 - 45 = 85
	def := &LotDefinitions[0] // fortune
	got := CalcLot(def, planets, asc, false, cusps)
	want := Normalize360(10 + 120 - 45)
	if math.Abs(got-want) > 0.01 {
		t.Errorf("Fortune nocturnal = %.2f, want %.2f", got, want)
	}
}
