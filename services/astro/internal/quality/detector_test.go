package quality

import "testing"

func TestDetectTechniques_BasicDetection(t *testing.T) {
	brief := `## SA
- SA Saturno ☌ MC (orb 0.4° +)
## DP
- Saturno ☌ MC — arco 45.2° (orb 0.8° apl)
## CONV
| Mar | SA+DP+Ecl | 82 |
## Ecl
- Eclipse solar ☌ Sol (orb 1.4°, mes 3)
`

	response := `Este año los arcos solares marcan un momento clave: Saturno
llega al MC con un orbe muy cerrado. Las direcciones primarias confirman
esta activación. La convergencia de marzo es notable, donde SA, DP y el
eclipse coinciden sobre puntos angulares.`

	result := DetectTechniques(brief, response)

	if len(result.Used) == 0 {
		t.Fatal("expected at least 1 technique detected as USED")
	}

	// SA should be detected
	foundSA := false
	for _, u := range result.Used {
		if u == "arcos_solares" {
			foundSA = true
		}
	}
	if !foundSA {
		t.Error("expected arcos_solares in Used, got:", result.Used)
	}

	// Convergencia should be detected
	foundConv := false
	for _, u := range result.Used {
		if u == "convergencia" {
			foundConv = true
		}
	}
	if !foundConv {
		t.Error("expected convergencia in Used, got:", result.Used)
	}

	t.Logf("Used: %v, Partial: %v, Omitted: %v", result.Used, result.Partial, result.Omitted)
}

func TestDetectTechniques_OmitsHighPriority(t *testing.T) {
	brief := `## SA
- SA Saturno ☌ MC (orb 0.4° +)
## DP
- Saturno ☌ MC — arco 45.2° (orb 0.8° apl)
`

	// Response mentions SA but NOT DP
	response := `El arco solar de Saturno sobre el MC es la activación principal del año.`

	result := DetectTechniques(brief, response)

	// DP should be OMITTED (weight >= 0.7, in brief, not in response)
	foundDPOmitted := false
	for _, o := range result.Omitted {
		if o == "direcciones_primarias" {
			foundDPOmitted = true
		}
	}
	if !foundDPOmitted {
		t.Error("expected direcciones_primarias in Omitted, got:", result.Omitted)
	}
}

func TestDetectTechniques_EmptyInputs(t *testing.T) {
	result := DetectTechniques("", "some response")
	if len(result.Used) != 0 || len(result.Omitted) != 0 {
		t.Error("empty brief should produce empty result")
	}

	result = DetectTechniques("## SA\n- SA Sol MC", "")
	if len(result.Used) != 0 || len(result.Omitted) != 0 {
		t.Error("empty response should produce empty result")
	}
}

func TestDetectTechniques_PartialForLowWeight(t *testing.T) {
	brief := `## Midpoints
- Sol/Luna = 15°22' Aries
`
	// Response doesn't mention midpoints
	response := `Los tránsitos lentos dominan el panorama anual.`

	result := DetectTechniques(brief, response)

	// Midpoints weight < 0.7, so should be PARTIAL not OMITTED
	foundPartial := false
	for _, p := range result.Partial {
		if p == "puntos_medios" {
			foundPartial = true
		}
	}
	if !foundPartial {
		t.Error("expected puntos_medios in Partial (low weight, not omitted), got Partial:", result.Partial, "Omitted:", result.Omitted)
	}
}
