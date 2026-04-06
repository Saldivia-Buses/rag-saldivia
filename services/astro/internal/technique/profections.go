package technique

import (
	"time"

	"github.com/Camionerou/rag-saldivia/services/astro/internal/astromath"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/natal"
)

// Profection holds profection data for a year.
type Profection struct {
	Age        int     `json:"age"`
	ActiveHouse int    `json:"active_house"` // 1-12
	ProfLon    float64 `json:"prof_lon"`     // profected ASC longitude
	ProfSign   string  `json:"prof_sign"`
	Lord       string  `json:"lord"`         // chronocrator (sign ruler)
	Theme      string  `json:"theme"`        // house theme
}

// ProfectionCascade holds the annual + monthly lord cascade.
type ProfectionCascade struct {
	YearLord   string         `json:"year_lord"`
	YearHouse  int            `json:"year_house"`
	MonthLords map[int]string `json:"month_lords"`  // 1-12
	MonthHouses map[int]int   `json:"month_houses"` // 1-12
}

// houseThemes maps house number to theme description (Spanish).
var houseThemes = map[int]string{
	1:  "identidad, cuerpo, inicio",
	2:  "recursos, finanzas, valores",
	3:  "comunicación, contratos, entorno inmediato",
	4:  "hogar, familia, raíces",
	5:  "creatividad, hijos, placer",
	6:  "salud, trabajo diario, servicio",
	7:  "relaciones, socios, contratos",
	8:  "crisis, transformación, recursos ajenos",
	9:  "viajes, estudios, filosofía",
	10: "carrera, reputación, autoridad",
	11: "amigos, grupos, proyectos",
	12: "retiro, espiritualidad, lo oculto",
}

// CalcProfection calculates profection data for a target year.
// birthDate is the native's birth date; chart provides ASC position.
// Age uses exact date calculation (matching Python's calculate_profection).
func CalcProfection(chart *natal.Chart, birthDate time.Time, targetYear int) *Profection {
	midYear := time.Date(targetYear, 7, 1, 0, 0, 0, 0, time.UTC)
	age := int(midYear.Sub(birthDate).Hours() / (24 * 365.25))

	// Active house: age mod 12, 1-indexed (age 0 = house 1)
	activeHouse := (age % 12) + 1

	// Profected ASC longitude: advance ASC by age signs (30° each)
	profLon := astromath.Normalize360(chart.ASC + float64(age)*30)
	signIdx := astromath.SignIndex(profLon)

	return &Profection{
		Age:         age,
		ActiveHouse: activeHouse,
		ProfLon:     profLon,
		ProfSign:    astromath.Signs[signIdx],
		Lord:        astromath.SignLord[signIdx],
		Theme:       houseThemes[activeHouse],
	}
}

// CalcProfectionCascade returns the annual lord + monthly breakdown.
// year_house = distance in signs from ASC sign to profected sign + 1.
func CalcProfectionCascade(chart *natal.Chart, birthDate time.Time, targetYear int) *ProfectionCascade {
	// Cascade uses simple year subtraction (matching Python's profection_lord_cascade)
	age := targetYear - birthDate.Year()
	profLon := astromath.Normalize360(chart.ASC + float64(age%12)*30)
	yearSignIdx := astromath.SignIndex(profLon)
	ascSignIdx := astromath.SignIndex(chart.ASC)
	yearHouse := ((yearSignIdx - ascSignIdx) % 12)
	if yearHouse < 0 {
		yearHouse += 12
	}
	yearHouse += 1

	monthLords := make(map[int]string)
	monthHouses := make(map[int]int)
	for m := 1; m <= 12; m++ {
		monthLon := astromath.Normalize360(profLon + float64(m-1)*30)
		mSignIdx := astromath.SignIndex(monthLon)
		monthLords[m] = astromath.SignLord[mSignIdx]
		mHouse := ((mSignIdx - ascSignIdx) % 12)
		if mHouse < 0 {
			mHouse += 12
		}
		monthHouses[m] = mHouse + 1
	}

	return &ProfectionCascade{
		YearLord:    astromath.SignLord[yearSignIdx],
		YearHouse:   yearHouse,
		MonthLords:  monthLords,
		MonthHouses: monthHouses,
	}
}
