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

func TestDomainRegistry_Plan13_TotalCount(t *testing.T) {
	domains := defaultDomains()
	if got := len(domains); got < 60 {
		t.Errorf("domain count = %d, want >= 60 (8 root + 50+ sub)", got)
	}
	t.Logf("total domains: %d", len(domains))
}

func TestDomainRegistry_Plan13_RootDomains(t *testing.T) {
	domains := defaultDomains()
	roots := make([]string, 0)
	for _, d := range domains {
		if d.Parent == "" {
			roots = append(roots, d.ID)
		}
	}

	expectedRoots := map[string]bool{
		"natal": true, "predictivo": true, "empresa": true,
		"electiva": true, "horaria": true, "relocation": true,
		"sinastria": true, "rectificacion": true,
	}
	for _, r := range roots {
		if !expectedRoots[r] {
			t.Errorf("unexpected root domain: %q", r)
		}
		delete(expectedRoots, r)
	}
	for missing := range expectedRoots {
		t.Errorf("missing expected root domain: %q", missing)
	}
}

func TestDomainRegistry_Plan13_NoDuplicateIDs(t *testing.T) {
	domains := defaultDomains()
	seen := make(map[string]bool, len(domains))
	for _, d := range domains {
		if seen[d.ID] {
			t.Errorf("duplicate domain ID: %q", d.ID)
		}
		seen[d.ID] = true
	}
}

func TestDomainRegistry_Plan13_UniqueKeywords(t *testing.T) {
	domains := defaultDomains()

	// Build parent map for ancestry check
	parentMap := make(map[string]string)
	for _, d := range domains {
		parentMap[d.ID] = d.Parent
	}
	isAncestor := func(child, ancestor string) bool {
		for cur := parentMap[child]; cur != ""; cur = parentMap[cur] {
			if cur == ancestor {
				return true
			}
		}
		return false
	}

	// Keyword collisions between parent and child are OK (child specializes parent).
	// Collisions between UNRELATED domains are real problems.
	kwMap := make(map[string]string)
	for _, d := range domains {
		for _, kw := range d.Keywords {
			if existing, ok := kwMap[kw]; ok {
				if isAncestor(d.ID, existing) || isAncestor(existing, d.ID) {
					continue // parent-child overlap is fine
				}
				t.Errorf("keyword %q used by unrelated domains %q and %q", kw, existing, d.ID)
			}
			kwMap[kw] = d.ID
		}
	}
}

func TestDomainRegistry_Plan13_AllResolveNoCycles(t *testing.T) {
	reg := mustRegistry(t)
	domains := defaultDomains()
	for _, d := range domains {
		resolved, err := reg.Resolve(d.ID)
		if err != nil {
			t.Errorf("Resolve(%q) failed: %v", d.ID, err)
			continue
		}
		if d.Parent != "" && len(resolved.InheritedFrom) < 2 {
			t.Errorf("Resolve(%q): subdomain should inherit, got InheritedFrom=%v", d.ID, resolved.InheritedFrom)
		}
	}
}

func TestDomainRegistry_Plan13_NewSubdomains(t *testing.T) {
	reg := mustRegistry(t)

	newDomains := []struct {
		id     string
		parent string
	}{
		// Natal subdomains
		{"personal", "natal"}, {"espiritualidad", "natal"}, {"viajes", "natal"},
		// Empresa subdomains
		{"negociacion", "empresa"}, {"deudores", "empresa"}, {"cash_flow", "empresa"},
		{"lanzamiento", "empresa"}, {"screening", "empresa"}, {"reunion_socios", "empresa"},
		{"sucesion", "empresa"}, {"sociedad", "empresa"}, {"timing", "empresa"},
		{"riesgos", "empresa"}, {"enterprise_general", "empresa"}, {"produccion", "empresa"},
		{"proveedores", "empresa"}, {"licitaciones", "empresa"}, {"expansion", "empresa"},
		{"calidad", "empresa"}, {"logistica", "empresa"}, {"empresa_competitivo", "empresa"},
		// Carrera subdomains
		{"educacion", "carrera"}, {"legal", "carrera"}, {"creatividad", "carrera"},
		{"fama", "carrera"}, {"relacion_laboral", "carrera"},
		// Salud subdomains
		{"crisis", "salud"}, {"deporte", "salud"}, {"cronico", "salud"},
		{"recuperacion", "salud"}, {"emergencia", "salud"},
		// Amor subdomains
		{"pareja", "amor"}, {"relacion", "amor"}, {"matrimonio", "amor"},
		// Dinero subdomains
		{"inmobiliario", "dinero"}, {"inversiones", "dinero"},
		{"deudas", "dinero"}, {"emprendimiento", "dinero"},
		// Familia subdomains
		{"herencia", "familia"}, {"fertilidad", "familia"}, {"hijos", "familia"},
		{"padres", "familia"}, {"divorcio", "familia"},
		// Competitivo subdomain
		{"competencia", "competitivo"},
		// Root additions
		{"sinastria", ""}, {"rectificacion", ""},
	}

	for _, tt := range newDomains {
		resolved, err := reg.Resolve(tt.id)
		if err != nil {
			t.Errorf("Resolve(%q): %v", tt.id, err)
			continue
		}
		if tt.parent != "" && resolved.Parent != tt.parent {
			t.Errorf("domain %q: parent = %q, want %q", tt.id, resolved.Parent, tt.parent)
		}
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
