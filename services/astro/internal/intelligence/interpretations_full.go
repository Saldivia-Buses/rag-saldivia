package intelligence

import "fmt"

// InterpretSA interprets a Solar Arc activation.
func InterpretSA(saplanet, aspect, natplanet string) string {
	nature := aspectNatureSp(aspect)
	return fmt.Sprintf("SA %s %s %s: %s de %s se dirige hacia %s — %s",
		saplanet, aspect, natplanet,
		planetKeywords[saplanet], saplanet, natplanet, nature)
}

// InterpretFirdaria interprets the active Firdaria period.
func InterpretFirdaria(majorLord, subLord string) string {
	mk := planetKeywords[majorLord]
	sk := planetKeywords[subLord]
	if sk == "" { sk = subLord }
	return fmt.Sprintf("Período Firdaria %s/%s: tema mayor de %s, matizado por %s",
		majorLord, subLord, mk, sk)
}

// InterpretProfection interprets the annual profection.
func InterpretProfection(house int, lord, theme string) string {
	return fmt.Sprintf("Casa %d profectada, cronócrata %s: año enfocado en %s",
		house, lord, theme)
}

// InterpretZR interprets Zodiacal Releasing periods.
func InterpretZR(lot, l1Lord, l2Lord string, loosing bool) string {
	base := fmt.Sprintf("ZR %s: L1=%s, L2=%s", lot, l1Lord, l2Lord)
	if loosing {
		base += " ⚠ LOOSING OF THE BOND — año de pico o crisis"
	}
	return base
}

// InterpretDirection interprets a Primary Direction.
func InterpretDirection(promissor, aspect, significator string, applying bool) string {
	dir := "separando"
	if applying { dir = "aplicando" }
	return fmt.Sprintf("DP %s %s %s (%s): activación de %s sobre %s",
		promissor, aspect, significator, dir,
		planetKeywords[promissor], significator)
}

// InterpretProgressions interprets secondary progressions.
func InterpretProgressions(planet, sign string, signIngress, houseIngress bool) string {
	base := fmt.Sprintf("%s progresado en %s", planet, sign)
	if signIngress {
		base += " — ⚠ INGRESO DE SIGNO: cambio de tono en " + planetKeywords[planet]
	}
	if houseIngress {
		base += " — ⚠ INGRESO DE CASA: nueva área de vida activada"
	}
	return base
}

// InterpretLunarReturn interprets a Lunar Return.
func InterpretLunarReturn(month int) string {
	return fmt.Sprintf("Retorno Lunar mes %d: reset emocional mensual, nueva semilla de 27 días", month)
}

// InterpretLunation interprets a New or Full Moon.
func InterpretLunation(typ, sign, natPoint, natAspect string, month int) string {
	base := fmt.Sprintf("Luna %s en %s (mes %d)", typ, sign, month)
	if natPoint != "" {
		base += fmt.Sprintf(" → %s %s natal: activación directa", natAspect, natPoint)
	}
	return base
}

// InterpretAntiscia interprets an antiscion activation.
func InterpretAntiscia(planetA, planetB string, orb float64, antiscionType string) string {
	return fmt.Sprintf("Antiscia %s: %s ↔ %s (orbe %.1f°) — conexión oculta por eje solsticial",
		antiscionType, planetA, planetB, orb)
}

// InterpretArabicPart interprets a lot/Arabic part.
func InterpretArabicPart(name, sign string, house int, description string) string {
	return fmt.Sprintf("Lote de %s en %s (casa %d): %s", name, sign, house, description)
}

// InterpretFixedStar interprets a fixed star conjunction.
func InterpretFixedStar(star, nature, natPoint string, orb float64) string {
	return fmt.Sprintf("★ %s conjunción %s (orbe %.1f°, naturaleza %s)",
		star, natPoint, orb, nature)
}

// InterpretCazimiCombustion interprets cazimi or combustion status.
func InterpretCazimiCombustion(planet, status string) string {
	switch status {
	case "cazimi":
		return fmt.Sprintf("%s en cazimi (0°17' del Sol): poder máximo, protección solar", planet)
	case "combust":
		return fmt.Sprintf("%s combusto (dentro de 8° del Sol): debilitado, oculto tras el Sol", planet)
	default:
		return ""
	}
}

func aspectNatureSp(aspect string) string {
	switch aspect {
	case "trine", "sextile", "trígono", "sextil":
		return "flujo armónico, oportunidad"
	case "square", "cuadratura":
		return "tensión productiva, acción necesaria"
	case "opposition", "oposición":
		return "polarización, necesidad de equilibrio"
	case "conjunction", "conjunción":
		return "fusión de energías, inicio de ciclo"
	default:
		return "activación"
	}
}
