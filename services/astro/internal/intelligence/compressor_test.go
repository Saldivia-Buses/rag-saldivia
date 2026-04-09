package intelligence

import "testing"

func TestCompressBCA_Reduction(t *testing.T) {
	// Sample brief mimicking real output from BuildBrief
	brief := `# BRIEF DE INTELIGENCIA ASTROLÓGICA — Adrián — 2026

## SEÑORES DEL TIEMPO

**Profección anual:** Casa 7 activa, signo Libra, cronócrata: Venus
  Tema de casa: Relaciones, socios, contratos

**Firdaria:** período mayor Saturno (27 años), sub-período Marte

**ZR Fortuna:** L1=Capricornio(Saturno), L2=Géminis(Mercurio)

## DIRECCIONES PRIMARIAS (precisión: meses)

- Saturno conjunción MC — arco 45.23° edad 45.9 (orbe 0.82° directo aplicando)
- Júpiter oposición ASC — arco 44.85° edad 45.5 (orbe 1.17° directo separando)
- Venus cuadratura Sol — arco 46.01° edad 46.7 (orbe 0.35° converso aplicando)
- Marte trígono Luna — arco 43.92° edad 44.6 (orbe 1.98° directo separando)

## ARCOS SOLARES

- SA Saturno conjunción MC (orbe 0.42° favorable)
- SA Plutón oposición Ascendente (orbe 0.85° desafiante)
- SA Júpiter trígono Sol (orbe 1.12° favorable)
- SA Venus cuadratura Luna (orbe 0.67° tenso)
- SA Neptuno conjunción Venus (orbe 1.34° mixto)

## PROGRESIONES SECUNDARIAS

- Sol prog en Géminis (casa 7)
- Luna prog en Escorpio (casa 12) ⚠ INGRESO de signo (Libra → Escorpio)
- Mercurio prog en Géminis (casa 7) Rx

## TRÁNSITOS LENTOS

- Saturno conjunción Luna (orbe 1.45° 3 pasadas mes 3 desafiante)
- Júpiter trígono Sol (orbe 2.10° 1 pasadas mes 8 favorable)
- Neptuno oposición Marte (orbe 0.89° 2 pasadas Rx mes 5 disolvente)
- Plutón cuadratura Venus (orbe 1.67° 2 pasadas mes 7 transformador)

## ESTACIONES PLANETARIAS

- Saturno SR en 15°23' Aries (mes 6, día 15)
- Júpiter SD en 08°45' Cáncer (mes 10, día 3) ⚠ cerca de Luna (orbe 2.1°)

## ECLIPSES

- Eclipse solar total → conjunción Sol (orbe 1.45°, mes 3)
- Eclipse lunar parcial → oposición Luna (orbe 2.80°, mes 9)

Sin activaciones de eclipses sobre puntos menores.

## REVOLUCIÓN SOLAR

ASC RS: 15°23' Géminis, MC RS: 08°45' Piscis

## DIGNIDADES Y DISPOSICIÓN

**Almutén Figuris:** Saturno (puntaje 23)
**Dispositor final:** Sol

## CONVERGENCIA TÉCNICA

| Mes | Técnicas | Score |
|-----|----------|-------|
| Mar | SA+DP+TR+Ecl | 82 |
| May | SA+Prog+Nep | 65 |
| Jul | DP+Plu+ZR | 58 |
| Oct | Jup+SD+Luna | 71 |
`

	compressed := CompressBCA(brief)

	// Should be significantly shorter
	ratio := float64(len(compressed)) / float64(len(brief))
	t.Logf("original: %d chars, compressed: %d chars, ratio: %.1f%%", len(brief), len(compressed), ratio*100)

	if ratio > 0.75 {
		t.Errorf("compression ratio %.1f%%, want < 75%% (at least 25%% reduction)", ratio*100)
	}

	// Must preserve key data (orbs compressed from 0.42 → 0.4)
	mustContain := []string{
		"0.4",  // SA orb (compressed from 0.42)
		"0.8",  // DP orb (compressed from 0.82)
		"Sat",  // planet abbreviated
		"MC",   // angle preserved
		"82",   // convergence score
	}
	for _, s := range mustContain {
		if !contains(compressed, s) {
			t.Errorf("compressed brief missing %q", s)
		}
	}

	// Must NOT contain verbose patterns
	mustNotContain := []string{
		"DIRECCIONES PRIMARIAS (precisión: meses)", // should be compacted to "## DP"
		"aplicando",                                 // should be "apl"
		"separando",                                 // should be "sep"
	}
	for _, s := range mustNotContain {
		if contains(compressed, s) {
			t.Errorf("compressed brief still contains verbose pattern %q", s)
		}
	}
}

func TestCompressBCA_ShortBrief(t *testing.T) {
	short := "Brief corto"
	if got := CompressBCA(short); got != short {
		t.Errorf("short brief should pass through unchanged, got %q", got)
	}
}

func TestCompressContext_RemovesEmpty(t *testing.T) {
	brief := "## ECLIPSES\n\nSin activaciones dentro del orbe.\n\n## ARCOS SOLARES\n\n- SA Sol MC\n"
	compressed := CompressContext(brief)
	if contains(compressed, "Sin activaciones") {
		t.Error("CompressContext should remove 'Sin activaciones' sections")
	}
	if !contains(compressed, "SA Sol MC") {
		t.Error("CompressContext should keep sections with data")
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && len(s) >= len(substr) && containsStr(s, substr)
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
