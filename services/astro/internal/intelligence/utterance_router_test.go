package intelligence

import "testing"

func TestUtteranceRouter_IndexBuilt(t *testing.T) {
	reg := mustRegistry(t)
	router := NewUtteranceRouter(reg)

	if len(router.index) == 0 {
		t.Fatal("utterance index is empty — JSON embed may have failed")
	}
	if router.totalUtterances < 1800 {
		t.Errorf("totalUtterances = %d, want >= 1800", router.totalUtterances)
	}
	if len(router.routes) < 50 {
		t.Errorf("routes = %d, want >= 50", len(router.routes))
	}
	t.Logf("index tokens: %d, routes: %d, utterances: %d", len(router.index), len(router.routes), router.totalUtterances)
}

func TestUtteranceRouter_BasicRouting(t *testing.T) {
	reg := mustRegistry(t)
	router := NewUtteranceRouter(reg)

	tests := []struct {
		query   string
		wantAny []string // acceptable domains
	}{
		{"como viene mi carrera este año", []string{"carrera"}},
		{"cuando me conviene operar", []string{"cirugia", "salud", "salud_cirugia"}},
		{"quiero invertir en acciones", []string{"inversiones", "dinero"}},
		{"como es la relacion con mi pareja", []string{"relacion", "amor", "pareja"}},
		{"tengo una licitacion importante", []string{"licitaciones", "empresa", "negociacion"}},
		{"fecha ideal para la boda", []string{"matrimonio", "amor", "electiva"}},
		{"mi padre esta enfermo", []string{"padres", "salud", "familia", "crisis"}},
		{"quiero rectificar mi hora de nacimiento", []string{"rectificacion"}},
	}

	for _, tt := range tests {
		intent := router.Parse(tt.query)
		found := false
		for _, want := range tt.wantAny {
			if intent.PrimaryDomain == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Parse(%q) = %q (conf=%.2f), want one of %v", tt.query, intent.PrimaryDomain, intent.Confidence, tt.wantAny)
		}
	}
}

func TestUtteranceRouter_FallbackToKeyword(t *testing.T) {
	reg := mustRegistry(t)
	router := NewUtteranceRouter(reg)

	// Very generic query that may not match utterances well
	intent := router.Parse("xyzabc random nonsense")
	if intent.PrimaryDomain == "" {
		t.Error("expected fallback domain, got empty")
	}
	// Should fall back to predictivo or some keyword match
	t.Logf("fallback domain: %s (conf=%.2f)", intent.PrimaryDomain, intent.Confidence)
}

func TestUtteranceRouter_QuickDomain(t *testing.T) {
	reg := mustRegistry(t)
	router := NewUtteranceRouter(reg)

	domain := router.QuickDomain("como viene el negocio este año")
	if domain == "" {
		t.Error("QuickDomain returned empty")
	}
	t.Logf("QuickDomain result: %s", domain)
}

func TestUtteranceRouter_ConfidenceRange(t *testing.T) {
	reg := mustRegistry(t)
	router := NewUtteranceRouter(reg)

	intent := router.Parse("quiero saber sobre mi carrera profesional")
	if intent.Confidence < 0 || intent.Confidence > 1 {
		t.Errorf("confidence = %.2f, want [0, 1]", intent.Confidence)
	}
}

func TestUtteranceRouter_FocusPoints(t *testing.T) {
	reg := mustRegistry(t)
	router := NewUtteranceRouter(reg)

	intent := router.Parse("como afecta saturno en mi casa 10")
	hasSaturn := false
	hasCasa := false
	for _, fp := range intent.FocusPoints {
		if fp == "saturno" {
			hasSaturn = true
		}
		if fp == "casa 10" {
			hasCasa = true
		}
	}
	if !hasSaturn {
		t.Error("expected saturno in FocusPoints")
	}
	if !hasCasa {
		t.Error("expected casa 10 in FocusPoints")
	}
}

func TestTokenize(t *testing.T) {
	tokens := tokenize("¿Cómo viene mi carrera este año?")
	// Should be lowercase, no accents, no stopwords, no punctuation
	for _, tok := range tokens {
		if tok == "como" || tok == "mi" || tok == "este" {
			t.Errorf("stopword %q should have been removed", tok)
		}
	}
	hasCarrera := false
	for _, tok := range tokens {
		if tok == "carrera" {
			hasCarrera = true
		}
	}
	if !hasCarrera {
		t.Error("expected 'carrera' in tokens")
	}
}
