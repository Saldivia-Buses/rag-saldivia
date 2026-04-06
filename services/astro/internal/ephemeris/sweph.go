// Package ephemeris wraps swephgo with an idiomatic Go API.
// All other packages call this — never swephgo directly.
//
// Design: swephgo already serializes all calls with an internal global mutex.
// CalcMu is an application-level mutex for compound operations where SetTopo +
// CalcPlanet must be atomic (e.g., BuildNatal with different observer locations).
package ephemeris

import (
	"fmt"
	"sync"

	swe "github.com/mshafiee/swephgo"
)

// Planet IDs — mirrors pyswisseph constants.
const (
	Sun      = 0
	Moon     = 1
	Mercury  = 2
	Venus    = 3
	Mars     = 4
	Jupiter  = 5
	Saturn   = 6
	Uranus   = 7
	Neptune  = 8
	Pluto    = 9
	TrueNode = 11
	MeanApog = 12 // Black Moon Lilith
	Chiron   = 15
	EclNutID = -1 // swe.SE_ECL_NUT
)

// Calc flags.
const (
	FlagSwieph     = 2     // swe.SEFLG_SWIEPH
	FlagSpeed      = 256   // swe.SEFLG_SPEED
	FlagEquatorial = 2048  // swe.SEFLG_EQUATORIAL
	FlagTopoctr    = 32768 // swe.SEFLG_TOPOCTR
)

// House systems.
const HouseTopocentric = 'T' // Polich-Page Topocentric

// Eclipse type flags.
const (
	EclTotal   = 4
	EclAnnular = 8
	EclPartial = 16
)

// PlanetPos holds calculated planetary position.
type PlanetPos struct {
	Lon   float64 // ecliptic longitude (degrees) — or RA if FlagEquatorial
	Lat   float64 // ecliptic latitude (degrees) — or Dec if FlagEquatorial
	Dist  float64 // distance (AU)
	Speed float64 // speed in longitude (degrees/day)
	RA    float64 // right ascension (degrees)
	Dec   float64 // declination (degrees)
}

// CalcMu protects compound operations where SetTopo + CalcPlanet must be atomic.
var CalcMu sync.Mutex

// Init sets the ephemeris data path. Must be called before any calculation.
func Init(ephePath string) {
	swe.SetEphePath([]byte(ephePath))
}

// Close releases ephemeris resources.
func Close() {
	swe.Close()
}

// SetTopo sets topocentric observer position.
// Caller must hold CalcMu when using this with subsequent CalcPlanet calls.
func SetTopo(geolon, geolat, altitude float64) {
	swe.SetTopo(geolon, geolat, altitude)
}

// JulDay converts a calendar date to Julian Day number (Gregorian calendar).
func JulDay(year, month, day int, hour float64) float64 {
	return swe.Julday(year, month, day, hour, 1)
}

// RevJul converts Julian Day back to calendar date.
func RevJul(jd float64) (year, month, day int, hour float64) {
	y := make([]int, 1)
	m := make([]int, 1)
	d := make([]int, 1)
	h := make([]float64, 1)
	swe.Revjul(jd, 1, y, m, d, h)
	return y[0], m[0], d[0], h[0]
}

// CalcPlanet calculates a planet's position for a given Julian Day (UT).
func CalcPlanet(jdUT float64, planet, flags int) (*PlanetPos, error) {
	xx := make([]float64, 6)
	serr := make([]byte, 256)

	ret := swe.CalcUt(jdUT, planet, flags, xx, serr)
	if ret < 0 {
		return nil, fmt.Errorf("swe.CalcUt(planet=%d): %s", planet, string(serr))
	}

	pos := &PlanetPos{
		Lon:   xx[0],
		Lat:   xx[1],
		Dist:  xx[2],
		Speed: xx[3],
	}
	if flags&FlagEquatorial != 0 {
		pos.RA = xx[0]
		pos.Dec = xx[1]
	}
	return pos, nil
}

// CalcPlanetFull calculates both ecliptic AND equatorial positions.
// Caller must hold CalcMu if using topocentric positions.
func CalcPlanetFull(jdUT float64, planet, baseFlags int) (*PlanetPos, error) {
	eclFlags := baseFlags &^ FlagEquatorial
	ecl, err := CalcPlanet(jdUT, planet, eclFlags)
	if err != nil {
		return nil, err
	}

	eqFlags := (baseFlags &^ FlagTopoctr) | FlagEquatorial | FlagSwieph | FlagSpeed
	eq, err := CalcPlanet(jdUT, planet, eqFlags)
	if err != nil {
		return nil, err
	}

	ecl.RA = eq.RA
	ecl.Dec = eq.Dec
	return ecl, nil
}

// CalcHouses calculates house cusps and angles.
// Returns cusps (13 elements: [0] unused, [1]-[12] are houses)
// and ascmc (10 elements: [0]=ASC, [1]=MC, [2]=ARMC, [3]=Vertex).
func CalcHouses(jdUT, geolat, geolon float64, hsys int) ([]float64, []float64, error) {
	cusps := make([]float64, 13)
	ascmc := make([]float64, 10)

	ret := swe.Houses(jdUT, geolat, geolon, hsys, cusps, ascmc)
	if ret < 0 {
		return nil, nil, fmt.Errorf("swe.Houses failed")
	}
	return cusps, ascmc, nil
}

// CalcHousesEx calculates house cusps with flags (e.g., topocentric).
func CalcHousesEx(jdUT float64, flags int, geolat, geolon float64, hsys int) ([]float64, []float64, error) {
	cusps := make([]float64, 13)
	ascmc := make([]float64, 10)

	ret := swe.HousesEx(jdUT, flags, geolat, geolon, hsys, cusps, ascmc)
	if ret < 0 {
		return nil, nil, fmt.Errorf("swe.HousesEx failed")
	}
	return cusps, ascmc, nil
}

// FixstarUT calculates a fixed star's position.
func FixstarUT(name string, jdUT float64, flags int) (float64, error) {
	xx := make([]float64, 6)
	serr := make([]byte, 256)
	star := make([]byte, len(name)+1)
	copy(star, name)

	ret := swe.FixstarUt(star, jdUT, flags, xx, serr)
	if ret < 0 {
		return 0, fmt.Errorf("swe.FixstarUt(%s): %s", name, string(serr))
	}
	return xx[0], nil
}

// SolEclipseWhenGlob finds next solar eclipse after tjdStart.
func SolEclipseWhenGlob(tjdStart float64, flags, eclType int) ([]float64, error) {
	tret := make([]float64, 10)
	serr := make([]byte, 256)

	ret := swe.SolEclipseWhenGlob(tjdStart, flags, eclType, tret, 0, serr)
	if ret < 0 {
		return nil, fmt.Errorf("swe.SolEclipseWhenGlob: %s", string(serr))
	}
	return tret, nil
}

// LunEclipseWhen finds next lunar eclipse after tjdStart.
func LunEclipseWhen(tjdStart float64, flags, eclType int) ([]float64, error) {
	tret := make([]float64, 10)
	serr := make([]byte, 256)

	ret := swe.LunEclipseWhen(tjdStart, flags, eclType, tret, 0, serr)
	if ret < 0 {
		return nil, fmt.Errorf("swe.LunEclipseWhen: %s", string(serr))
	}
	return tret, nil
}

// GetAyanamsaUT returns the ayanamsa (sidereal offset) for a Julian Day.
func GetAyanamsaUT(jdUT float64) float64 {
	return swe.GetAyanamsaUt(jdUT)
}

// EclNut calculates obliquity of the ecliptic.
func EclNut(jdUT float64) (float64, error) {
	xx := make([]float64, 6)
	serr := make([]byte, 256)

	ret := swe.CalcUt(jdUT, EclNutID, FlagSwieph, xx, serr)
	if ret < 0 {
		return 0, fmt.Errorf("swe.CalcUt(ECL_NUT): %s", string(serr))
	}
	return xx[0], nil
}

// RiseTrans calculates rise/set times.
func RiseTrans(jdUT float64, planet int, epheflag, rsmi int, geopos []float64, atpress, attemp float64) (float64, error) {
	tret := make([]float64, 1)
	serr := make([]byte, 256)

	ret := swe.RiseTrans(jdUT, planet, nil, epheflag, rsmi, geopos, atpress, attemp, tret, serr)
	if ret < 0 {
		return 0, fmt.Errorf("swe.RiseTrans: %s", string(serr))
	}
	return tret[0], nil
}
