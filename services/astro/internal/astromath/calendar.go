package astromath

// ArgentineHoliday represents a national holiday.
type ArgentineHoliday struct {
	Date string // MM-DD
	Name string
}

// ArgentineHolidays2026 contains fixed Argentine holidays for 2026.
// Injected into timing context to avoid recommending meetings on holidays.
var ArgentineHolidays2026 = []ArgentineHoliday{
	{"01-01", "Año Nuevo"},
	{"02-16", "Carnaval"},
	{"02-17", "Carnaval"},
	{"03-24", "Día de la Memoria"},
	{"04-02", "Día de Malvinas"},
	{"04-03", "Viernes Santo"},
	{"05-01", "Día del Trabajador"},
	{"05-25", "Revolución de Mayo"},
	{"06-15", "Paso a la Inmortalidad de Güemes"},
	{"06-20", "Paso a la Inmortalidad de Belgrano"},
	{"07-09", "Día de la Independencia"},
	{"08-17", "Paso a la Inmortalidad de San Martín"},
	{"10-12", "Día del Respeto a la Diversidad Cultural"},
	{"11-20", "Día de la Soberanía Nacional"},
	{"12-08", "Inmaculada Concepción"},
	{"12-25", "Navidad"},
}

// IsHoliday checks if a given month/day is an Argentine holiday.
func IsHoliday(month, day int) bool {
	key := ""
	if month < 10 {
		key = "0"
	}
	key += string(rune('0'+month/10)) + string(rune('0'+month%10)) + "-"
	if day < 10 {
		key += "0"
	}
	key += string(rune('0'+day/10)) + string(rune('0'+day%10))

	for _, h := range ArgentineHolidays2026 {
		if h.Date == key {
			return true
		}
	}
	return false
}

// GetMonthHolidays returns holidays for a specific month.
func GetMonthHolidays(month int) []ArgentineHoliday {
	prefix := ""
	if month < 10 {
		prefix = "0"
	}
	prefix += string(rune('0'+month/10)) + string(rune('0'+month%10))

	var holidays []ArgentineHoliday
	for _, h := range ArgentineHolidays2026 {
		if len(h.Date) >= 2 && h.Date[:2] == prefix {
			holidays = append(holidays, h)
		}
	}
	return holidays
}
