package astromath

// DecumbituraResult holds medical astrology analysis.
type DecumbituraResult struct {
	CriticalDays []CriticalDay `json:"critical_days"`
	Vulnerable   []string      `json:"vulnerable_zones"` // body zones with affliction
}

// CriticalDay marks a day when the Moon forms a hard aspect to its natal position.
type CriticalDay struct {
	Day       int    `json:"day"`     // day number from illness onset (1-based)
	Aspect    string `json:"aspect"`  // "cuadratura", "oposición", "conjunción"
	LunarAge  float64 `json:"lunar_age"` // days into lunar cycle
	Severity  string `json:"severity"` // "alta", "media"
}

// CalcCriticalDays computes the classical critical days of an illness
// based on the Moon's synodic cycle (~29.5 days).
// In decumbitura tradition, the crisis points are:
//   Day 7  (first quarter, square)
//   Day 14 (full moon, opposition)
//   Day 21 (last quarter, square)
//   Day 28 (new moon, conjunction)
// These are approximate — exact timing depends on the Moon's speed.
func CalcCriticalDays() *DecumbituraResult {
	lunarCycle := 29.53059 // synodic month in days

	days := []CriticalDay{
		{Day: 7, Aspect: "cuadratura", LunarAge: lunarCycle / 4, Severity: "alta"},
		{Day: 14, Aspect: "oposición", LunarAge: lunarCycle / 2, Severity: "alta"},
		{Day: 21, Aspect: "cuadratura", LunarAge: lunarCycle * 3 / 4, Severity: "media"},
		{Day: 28, Aspect: "conjunción", LunarAge: lunarCycle, Severity: "alta"},
		// Extended critical days (second cycle)
		{Day: 35, Aspect: "cuadratura", LunarAge: lunarCycle + lunarCycle/4, Severity: "media"},
		{Day: 42, Aspect: "oposición", LunarAge: lunarCycle + lunarCycle/2, Severity: "media"},
	}

	return &DecumbituraResult{
		CriticalDays: days,
	}
}
