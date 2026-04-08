package intelligence

import "testing"

func TestNewDomainRegistry(t *testing.T) {
	domains := defaultDomains()
	reg, err := NewDomainRegistry(domains)
	if err != nil {
		t.Fatalf("NewDomainRegistry: %v", err)
	}
	if len(reg.RootDomains()) == 0 {
		t.Error("no root domains")
	}
	if len(reg.AllIDs()) < 12 {
		t.Errorf("expected at least 12 domains, got %d", len(reg.AllIDs()))
	}
}

func TestResolve_Root(t *testing.T) {
	reg := mustRegistry(t)
	d, err := reg.Resolve("natal")
	if err != nil {
		t.Fatalf("Resolve natal: %v", err)
	}
	if d.ID != "natal" {
		t.Errorf("ID = %q, want natal", d.ID)
	}
	if len(d.TechniquesRequired) == 0 {
		t.Error("natal should have required techniques")
	}
	if len(d.InheritedFrom) != 1 || d.InheritedFrom[0] != "natal" {
		t.Errorf("inherited = %v, want [natal]", d.InheritedFrom)
	}
}

func TestResolve_Subdomain_Inheritance(t *testing.T) {
	reg := mustRegistry(t)
	d, err := reg.Resolve("carrera")
	if err != nil {
		t.Fatalf("Resolve carrera: %v", err)
	}
	if d.Parent != "predictivo" {
		t.Errorf("carrera parent = %q, want predictivo", d.Parent)
	}
	// Should inherit predictivo's required techniques
	hasTransitos := false
	for _, tech := range d.TechniquesRequired {
		if tech == TechTransits {
			hasTransitos = true
		}
	}
	if !hasTransitos {
		t.Error("carrera should inherit transitos from predictivo")
	}
	// Should have its own required (profecciones)
	hasProfecciones := false
	for _, tech := range d.TechniquesRequired {
		if tech == TechProfections {
			hasProfecciones = true
		}
	}
	if !hasProfecciones {
		t.Error("carrera should require profecciones")
	}
	// InheritedFrom should show chain
	if len(d.InheritedFrom) < 2 {
		t.Errorf("InheritedFrom = %v, want [carrera, predictivo]", d.InheritedFrom)
	}
}

func TestResolve_Subdomain_Precautions(t *testing.T) {
	reg := mustRegistry(t)
	d, err := reg.Resolve("salud")
	if err != nil {
		t.Fatalf("Resolve salud: %v", err)
	}
	if len(d.Precautions) == 0 {
		t.Error("salud should have precautions")
	}
	// Check a specific precaution exists
	found := false
	for _, p := range d.Precautions {
		if p == "NUNCA diagnosticar enfermedades — solo tendencias energéticas" {
			found = true
		}
	}
	if !found {
		t.Error("salud missing expected precaution")
	}
}

func TestResolve_Fallback(t *testing.T) {
	reg := mustRegistry(t)
	d, err := reg.Resolve("nonexistent_domain")
	if err != nil {
		t.Fatalf("Resolve fallback: %v", err)
	}
	if d.ID != "predictivo" {
		t.Errorf("fallback ID = %q, want predictivo", d.ID)
	}
}

func mustRegistry(t *testing.T) *DomainRegistry {
	t.Helper()
	reg, err := NewDomainRegistry(defaultDomains())
	if err != nil {
		t.Fatalf("NewDomainRegistry: %v", err)
	}
	return reg
}
