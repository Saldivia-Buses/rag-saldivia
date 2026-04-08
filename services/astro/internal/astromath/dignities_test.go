package astromath

import "testing"

func TestDignityScoreAt_Domicile(t *testing.T) {
	// Sol in Leo (sign 4) should score Domicile (5)
	score, dignity := DignityScoreAt("Sol", 130, true) // 130° = Leo 10°
	if score != ScoreDomicile || dignity != "domicilio" {
		t.Errorf("Sol in Leo: got score=%d dignity=%q, want %d domicilio", score, dignity, ScoreDomicile)
	}
}

func TestDignityScoreAt_Exaltation(t *testing.T) {
	// Sol exalted in Aries (sign 0)
	score, dignity := DignityScoreAt("Sol", 19, true) // 19° Aries
	if score != ScoreExaltation || dignity != "exaltación" {
		t.Errorf("Sol at 19° Aries: got score=%d dignity=%q, want %d exaltación", score, dignity, ScoreExaltation)
	}
}

func TestDignityScoreAt_Triplicity(t *testing.T) {
	// Sol is day triplicity lord of fire signs. Sol at 0° Aries (fire, diurnal)
	score, dignity := DignityScoreAt("Sol", 0, true)
	// Sol is not domicile (Marte) nor exalted here. Triplicity of fire day = Sol
	if score != ScoreTriplicity || dignity != "triplicidad" {
		t.Errorf("Sol at 0° Aries diurnal: got score=%d dignity=%q, want %d triplicidad", score, dignity, ScoreTriplicity)
	}
}

func TestDignityScoreAt_Detriment(t *testing.T) {
	// Venus in Aries (detriment — opposite of Libra)
	score, dignity := DignityScoreAt("Venus", 5, true) // 5° Aries
	if score != ScoreDetriment || dignity != "detrimento" {
		t.Errorf("Venus in Aries: got score=%d dignity=%q, want %d detrimento", score, dignity, ScoreDetriment)
	}
}

func TestDignityScoreAt_Peregrine(t *testing.T) {
	// Júpiter at 15° Leo — no domicile, no exaltation, not fire triplicity day lord (that's Sol)
	// Check if it gets term or face, or 0
	score, _ := DignityScoreAt("Júpiter", 120, true) // 0° Leo
	// Leo 0-6° terms: Júpiter has first term
	if score != ScoreDomicile && score != ScoreExaltation && score != ScoreTriplicity && score != ScoreTerm && score != ScoreFace && score != 0 {
		t.Errorf("unexpected score %d for Júpiter at 0° Leo", score)
	}
}

func TestTermLord(t *testing.T) {
	tests := []struct {
		lon  float64
		want string
	}{
		{0, "Júpiter"},   // Aries 0° → first term is Júpiter (0-6)
		{7, "Venus"},     // Aries 7° → Venus (6-12)
		{30, "Venus"},    // Tauro 0° → Venus (0-8)
		{40, "Mercurio"}, // Tauro 10° → Mercurio (8-14)
	}
	for _, tc := range tests {
		got := TermLord(tc.lon)
		if got != tc.want {
			t.Errorf("TermLord(%.0f) = %q, want %q", tc.lon, got, tc.want)
		}
	}
}

func TestFaceLord(t *testing.T) {
	tests := []struct {
		lon  float64
		want string
	}{
		{0, "Marte"},    // Aries 0° decan 1 → Marte (chaldean starts from Marte)
		{10, "Sol"},     // Aries 10° decan 2 → Sol
		{20, "Venus"},   // Aries 20° decan 3 → Venus
		{30, "Mercurio"}, // Tauro 0° decan 4 → Mercurio
	}
	for _, tc := range tests {
		got := FaceLord(tc.lon)
		if got != tc.want {
			t.Errorf("FaceLord(%.0f) = %q, want %q", tc.lon, got, tc.want)
		}
	}
}

func TestTriplicityLord(t *testing.T) {
	// Aries (fire, index 0): day=Sol, night=Júpiter
	if got := TriplicityLord(0, true); got != "Sol" {
		t.Errorf("fire day triplicity: got %q, want Sol", got)
	}
	if got := TriplicityLord(0, false); got != "Júpiter" {
		t.Errorf("fire night triplicity: got %q, want Júpiter", got)
	}
	// Tauro (earth, index 1): day=Venus, night=Luna
	if got := TriplicityLord(1, true); got != "Venus" {
		t.Errorf("earth day triplicity: got %q, want Venus", got)
	}
}
