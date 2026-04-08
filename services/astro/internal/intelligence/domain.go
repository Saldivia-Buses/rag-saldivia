package intelligence

import "fmt"

// TechniqueWeight pairs a technique ID with importance weight and focus area.
type TechniqueWeight struct {
	ID     string  // e.g. "transitos", "profecciones"
	Weight float64 // 0.0-1.0
	Focus  string  // optional: "casa_10", "casa_7", etc.
}

// Domain defines a query domain with technique requirements.
type Domain struct {
	ID                 string
	Name               string // display name (Spanish)
	Parent             string // "" for root domains
	TechniquesRequired []string
	TechniquesExpected []string
	TechniquesBrief    []TechniqueWeight
	Precautions        []string
	Keywords           []string // for intent matching
}

// ResolvedDomain is a Domain with inheritance fully applied.
type ResolvedDomain struct {
	Domain
	InheritedFrom []string
}

// DomainRegistry holds all domains and provides lookup/resolution.
type DomainRegistry struct {
	domains map[string]*Domain
	roots   []string
	aliases map[string]string
}

// NewDomainRegistry creates a registry from domain definitions.
func NewDomainRegistry(domains []Domain) (*DomainRegistry, error) {
	r := &DomainRegistry{
		domains: make(map[string]*Domain, len(domains)),
		aliases: make(map[string]string),
	}
	for i := range domains {
		d := &domains[i]
		if _, exists := r.domains[d.ID]; exists {
			return nil, fmt.Errorf("duplicate domain ID: %s", d.ID)
		}
		r.domains[d.ID] = d
		if d.Parent == "" {
			r.roots = append(r.roots, d.ID)
		}
	}
	// Validate parent references
	for _, d := range r.domains {
		if d.Parent != "" {
			if _, ok := r.domains[d.Parent]; !ok {
				return nil, fmt.Errorf("domain %q references unknown parent %q", d.ID, d.Parent)
			}
		}
	}
	return r, nil
}

// Resolve resolves a domain ID with full inheritance applied.
// Detects cycles to prevent stack overflow (D6 fix).
func (r *DomainRegistry) Resolve(domainID string) (*ResolvedDomain, error) {
	return r.resolveWithVisited(domainID, make(map[string]bool))
}

func (r *DomainRegistry) resolveWithVisited(domainID string, visited map[string]bool) (*ResolvedDomain, error) {
	if visited[domainID] {
		return nil, fmt.Errorf("cycle detected in domain inheritance: %s", domainID)
	}
	visited[domainID] = true
	// Check alias
	if canonical, ok := r.aliases[domainID]; ok {
		domainID = canonical
	}
	d, ok := r.domains[domainID]
	if !ok {
		// Fallback to predictivo
		d, ok = r.domains["predictivo"]
		if !ok {
			return nil, fmt.Errorf("domain %q not found and no fallback", domainID)
		}
	}

	if d.Parent == "" {
		return &ResolvedDomain{Domain: *d, InheritedFrom: []string{d.ID}}, nil
	}

	// Apply inheritance
	parent, err := r.resolveWithVisited(d.Parent, visited)
	if err != nil {
		return nil, err
	}

	merged := ResolvedDomain{
		Domain:        *d,
		InheritedFrom: append([]string{d.ID}, parent.InheritedFrom...),
	}

	// Merge required: parent + child (deduplicated)
	seen := make(map[string]bool)
	for _, t := range parent.TechniquesRequired {
		if !seen[t] {
			merged.TechniquesRequired = append(merged.TechniquesRequired, t)
			seen[t] = true
		}
	}
	for _, t := range d.TechniquesRequired {
		if !seen[t] {
			merged.TechniquesRequired = append(merged.TechniquesRequired, t)
			seen[t] = true
		}
	}

	// Merge expected
	seen = make(map[string]bool)
	for _, t := range parent.TechniquesExpected {
		seen[t] = true
	}
	for _, t := range d.TechniquesExpected {
		seen[t] = true
	}
	merged.TechniquesExpected = make([]string, 0, len(seen))
	for t := range seen {
		merged.TechniquesExpected = append(merged.TechniquesExpected, t)
	}

	// Brief: child overrides parent (child weights take precedence)
	if len(d.TechniquesBrief) == 0 {
		merged.TechniquesBrief = parent.TechniquesBrief
	}

	// Precautions: parent + child (deduplicated)
	precSeen := make(map[string]bool)
	for _, p := range parent.Precautions {
		if !precSeen[p] {
			merged.Precautions = append(merged.Precautions, p)
			precSeen[p] = true
		}
	}
	for _, p := range d.Precautions {
		if !precSeen[p] {
			merged.Precautions = append(merged.Precautions, p)
			precSeen[p] = true
		}
	}

	// Keywords: child only (more specific)
	if len(d.Keywords) == 0 {
		merged.Keywords = parent.Keywords
	}

	return &merged, nil
}

// Get returns a domain by ID without inheritance.
func (r *DomainRegistry) Get(id string) *Domain {
	if canonical, ok := r.aliases[id]; ok {
		id = canonical
	}
	return r.domains[id]
}

// RootDomains returns root domain IDs.
func (r *DomainRegistry) RootDomains() []string { return r.roots }

// AllIDs returns all domain IDs.
func (r *DomainRegistry) AllIDs() []string {
	ids := make([]string, 0, len(r.domains))
	for id := range r.domains {
		ids = append(ids, id)
	}
	return ids
}

// Technique ID constants.
const (
	TechNatal         = "natal"
	TechTransits      = "transitos"
	TechSolarArc      = "arcos_solares"
	TechPrimaryDir    = "direcciones_primarias"
	TechProgressions  = "progresiones"
	TechSolarReturn   = "revolucion_solar"
	TechLunarReturn   = "retorno_lunar"
	TechProfections   = "profecciones"
	TechFirdaria      = "firdaria"
	TechEclipses      = "eclipses"
	TechFixedStars    = "estrellas_fijas"
	TechZR            = "zodiacal_releasing"
	TechStations      = "estaciones"
	TechConvergence   = "convergencia"
	TechAlmuten       = "almuten"
	TechLots          = "lotes"
	TechDisposition   = "disposicion"
	TechSect          = "secta"
	TechDecennials    = "deceniales"
	TechTertiaryProg  = "progresiones_terciarias"
	TechFastTransits  = "transitos_rapidos"
	TechLunations     = "lunaciones"
	TechPrenatalEcl   = "eclipse_prenatal"
	TechEclTriggers   = "eclipse_triggers"
	TechPlanetCycles  = "ciclos_planetarios"
	TechMidpoints     = "puntos_medios"
	TechDeclinations  = "declinaciones"
	TechActivChains   = "cadenas_activacion"
	TechTimingWindows = "ventanas_timing"
	TechTemperament   = "temperamento"
	TechMelothesia    = "melotesia"
	TechHyleg         = "hyleg"
	TechSynastry      = "sinastria"
	TechComposite     = "compuesta"
	TechElectional    = "electiva"
	TechHorary        = "horaria"
	TechRectification = "rectificacion"
	TechACG           = "astrocartografia"
	TechMercuryRx     = "mercurio_rx"
	TechChiron        = "quiron"
	TechNodes         = "nodos"
	TechDivisor       = "divisor"
	TechTriplicity    = "triplicidad"
	TechRelocation    = "relocacion"
	TechCorpHouses    = "casas_corporativas"
	TechPlanetReturns = "retornos_planetarios"
	TechVocational    = "vocacional"
	TechSabian        = "sabiano"
	TechLilith        = "lilith"
	TechVertex        = "vertex"
	TechAntiscia      = "antiscia"
	TechDecumbitura   = "decumbitura"
	TechChronocrator  = "cronocrator"
	TechDavison       = "davison"
	TechWeeklyTR      = "transitos_semanales"
	TechAriesIngress  = "ingreso_aries"
)
