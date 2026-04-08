package intelligence

import "testing"

func TestDetectFollowUp_ContinuationPatterns(t *testing.T) {
	tests := []struct {
		msg  string
		want bool
	}{
		// Follow-ups (should detect)
		{"Y en mayo?", true},
		{"Pero que pasa con Jupiter?", true},
		{"Contame mas sobre eso", true},
		{"En que mes es mejor?", true},
		{"Y eso que significa?", true},
		{"También quiero saber sobre el amor", true},
		{"Podes ampliar sobre marzo?", true},
		{"Y en septiembre?", true},

		// NOT follow-ups (new consultations)
		{"Carta de Juan Garcia", false},
		{"Naci el 15 de marzo de 1990 a las 14:30 en Buenos Aires", false},
		{"Quiero saber la carta natal de mi hijo nacido en 1985", false},
		{"Amor 2027", false},           // explicit domain + year = new query (H5)
		{"Jupiter 2027?", false},       // planet + year + ? = new query, NOT follow-up (H5)
		// Long message (>25 words) — not a follow-up
		{"Quiero que me hagas un analisis completo y detallado de todas las tecnicas astrologicas disponibles para este año incluyendo transitos progresiones y todo lo demas", false},
	}

	for _, tt := range tests {
		result := DetectFollowUp(tt.msg, true, "predictivo", "contact-123")
		got := result != nil
		if got != tt.want {
			t.Errorf("DetectFollowUp(%q) = %v, want %v", tt.msg, got, tt.want)
		}
	}
}

func TestDetectFollowUp_NoHistory(t *testing.T) {
	result := DetectFollowUp("Y en mayo?", false, "", "")
	if result != nil {
		t.Error("expected nil when hasHistory=false")
	}
}

func TestDetectFollowUp_InheritedContext(t *testing.T) {
	result := DetectFollowUp("Y en mayo?", true, "carrera", "contact-456")
	if result == nil {
		t.Fatal("expected follow-up detection")
	}
	if result.DomainID != "carrera" {
		t.Errorf("DomainID = %q, want carrera", result.DomainID)
	}
	if result.ContactID != "contact-456" {
		t.Errorf("ContactID = %q, want contact-456", result.ContactID)
	}
}

func TestDetectFollowUp_ShortQuestion(t *testing.T) {
	// Very short questions ending in ? should be follow-ups
	result := DetectFollowUp("Y marte?", true, "predictivo", "c1")
	if result == nil {
		t.Error("short question 'Y marte?' should be detected as follow-up")
	}
}

func TestDetectFollowUp_EmptyMessage(t *testing.T) {
	result := DetectFollowUp("", true, "predictivo", "c1")
	if result != nil {
		t.Error("empty message should not be a follow-up")
	}
}
