package intelligence

import "testing"

func TestParseIntent_Career(t *testing.T) {
	reg := mustRegistry(t)
	intent := ParseIntent("como le va a Adrian en el trabajo este año", reg)
	if intent.PrimaryDomain != "carrera" {
		t.Errorf("domain = %q, want carrera", intent.PrimaryDomain)
	}
	if intent.Confidence <= 0 {
		t.Error("confidence should be > 0")
	}
}

func TestParseIntent_Health(t *testing.T) {
	reg := mustRegistry(t)
	intent := ParseIntent("tiene algún problema de salud este año?", reg)
	if intent.PrimaryDomain != "salud" {
		t.Errorf("domain = %q, want salud", intent.PrimaryDomain)
	}
}

func TestParseIntent_Love(t *testing.T) {
	reg := mustRegistry(t)
	intent := ParseIntent("como viene el amor y las relaciones para Maria", reg)
	if intent.PrimaryDomain != "amor" {
		t.Errorf("domain = %q, want amor", intent.PrimaryDomain)
	}
}

func TestParseIntent_Enterprise(t *testing.T) {
	reg := mustRegistry(t)
	intent := ParseIntent("como le va a la empresa en ventas", reg)
	// Could match empresa or empresa_ventas
	if intent.PrimaryDomain != "empresa" && intent.PrimaryDomain != "empresa_ventas" {
		t.Errorf("domain = %q, want empresa or empresa_ventas", intent.PrimaryDomain)
	}
}

func TestParseIntent_Electional(t *testing.T) {
	reg := mustRegistry(t)
	intent := ParseIntent("cual es la mejor fecha para firmar el contrato", reg)
	if intent.PrimaryDomain != "electiva" {
		t.Errorf("domain = %q, want electiva", intent.PrimaryDomain)
	}
}

func TestParseIntent_Fallback(t *testing.T) {
	reg := mustRegistry(t)
	intent := ParseIntent("hola como estas", reg)
	if intent.PrimaryDomain != "predictivo" {
		t.Errorf("domain = %q, want predictivo (fallback)", intent.PrimaryDomain)
	}
	if intent.Confidence > 0.4 {
		t.Errorf("fallback confidence = %.2f, should be low", intent.Confidence)
	}
}

func TestParseIntent_FocusPoints(t *testing.T) {
	reg := mustRegistry(t)
	intent := ParseIntent("Saturno en casa 10 que significa para mi carrera", reg)
	hasSaturno := false
	hasCasa10 := false
	for _, fp := range intent.FocusPoints {
		if fp == "saturno" {
			hasSaturno = true
		}
		if fp == "casa 10" {
			hasCasa10 = true
		}
	}
	if !hasSaturno {
		t.Error("should detect Saturno as focus point")
	}
	if !hasCasa10 {
		t.Error("should detect casa 10 as focus point")
	}
}

func TestParseIntent_Money(t *testing.T) {
	reg := mustRegistry(t)
	intent := ParseIntent("como estan mis finanzas, puedo invertir?", reg)
	// "invertir" may match subdomain "inversiones" (child of dinero) — both are acceptable
	if intent.PrimaryDomain != "dinero" && intent.PrimaryDomain != "inversiones" {
		t.Errorf("domain = %q, want dinero or inversiones", intent.PrimaryDomain)
	}
}
