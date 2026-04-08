package intelligence

import (
	"fmt"
	"strings"
)

// ResponseSkeleton generates a structured template that guides the LLM response format.
// Ensures consistent output across different domains and query types.
type ResponseSkeleton struct {
	Title    string            `json:"title"`
	Sections []SkeletonSection `json:"sections"`
}

// SkeletonSection is one section in the response template.
type SkeletonSection struct {
	Header      string `json:"header"`
	Instruction string `json:"instruction"` // what to write here
	Required    bool   `json:"required"`
}

// BuildSkeleton creates a response skeleton based on domain and cross-references.
func BuildSkeleton(domain *ResolvedDomain, crossRefCount int) *ResponseSkeleton {
	skeleton := &ResponseSkeleton{
		Title: domain.Name,
	}

	// Always start with convergences if any exist
	if crossRefCount > 0 {
		skeleton.Sections = append(skeleton.Sections, SkeletonSection{
			Header:      "Convergencias clave",
			Instruction: "Las 2-3 convergencias más significativas entre técnicas. Tejidas narrativamente.",
			Required:    true,
		})
	}

	// Domain-specific sections
	switch domain.ID {
	case "natal":
		skeleton.Sections = append(skeleton.Sections,
			SkeletonSection{Header: "Personalidad", Instruction: "Sol, Luna, ASC — la tríada básica.", Required: true},
			SkeletonSection{Header: "Dignidades", Instruction: "Almutén, secta, dispositor final.", Required: false},
			SkeletonSection{Header: "Puntos sensibles", Instruction: "Lotes, estrellas fijas relevantes.", Required: false},
		)
	case "predictivo", "carrera", "salud", "amor", "dinero", "familia":
		skeleton.Sections = append(skeleton.Sections,
			SkeletonSection{Header: "Señores del tiempo", Instruction: "Profección + firdaria + ZR — quién rige el año.", Required: true},
			SkeletonSection{Header: "Activaciones", Instruction: "SA + PD + tránsitos con timing mensual.", Required: true},
			SkeletonSection{Header: "Ventanas", Instruction: "Meses de mayor actividad y su naturaleza.", Required: true},
			SkeletonSection{Header: "Recomendación", Instruction: "Acción concreta con fecha sugerida.", Required: true},
		)
	case "empresa":
		skeleton.Sections = append(skeleton.Sections,
			SkeletonSection{Header: "Panorama del período", Instruction: "Outlook general para el negocio.", Required: true},
			SkeletonSection{Header: "Timing de negocios", Instruction: "Ventanas favorables con fechas.", Required: true},
			SkeletonSection{Header: "Riesgos", Instruction: "Meses a cuidar y en qué área.", Required: true},
			SkeletonSection{Header: "Acciones", Instruction: "Recomendaciones concretas por prioridad.", Required: true},
		)
	case "electiva":
		skeleton.Sections = append(skeleton.Sections,
			SkeletonSection{Header: "Fechas recomendadas", Instruction: "Top 3-5 fechas con razones.", Required: true},
			SkeletonSection{Header: "Fechas a evitar", Instruction: "Períodos desfavorables.", Required: true},
		)
	case "horaria":
		skeleton.Sections = append(skeleton.Sections,
			SkeletonSection{Header: "Radicalidad", Instruction: "¿Es la carta válida para juicio?", Required: true},
			SkeletonSection{Header: "Juicio", Instruction: "Respuesta basada en significadores.", Required: true},
			SkeletonSection{Header: "Timing", Instruction: "Cuándo se manifiesta el resultado.", Required: false},
		)
	}

	return skeleton
}

// FormatSkeletonGuide produces text for the system prompt.
func FormatSkeletonGuide(skeleton *ResponseSkeleton) string {
	var b strings.Builder
	b.WriteString("## ESTRUCTURA DE RESPUESTA\n\n")
	for i, s := range skeleton.Sections {
		req := ""
		if s.Required {
			req = " (OBLIGATORIO)"
		}
		b.WriteString(fmt.Sprintf("%d. **%s**%s: %s\n", i+1, s.Header, req, s.Instruction))
	}
	return b.String()
}
