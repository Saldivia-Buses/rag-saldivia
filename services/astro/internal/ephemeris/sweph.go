// Package ephemeris wraps swephgo with an idiomatic Go API.
// All other packages call this — never swephgo directly.
package ephemeris

// Init sets the ephemeris data path. Must be called before any calculation.
func Init(ephePath string) {
	// Will be implemented in Phase 2
}

// Close releases ephemeris resources.
func Close() {
	// Will be implemented in Phase 2
}
