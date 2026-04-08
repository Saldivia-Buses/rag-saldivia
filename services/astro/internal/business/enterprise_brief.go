package business

import (
	"fmt"
	"strings"

	"github.com/Camionerou/rag-saldivia/services/astro/internal/natal"
)

// BuildEnterpriseBrief creates a 2-layer brief for enterprise queries.
// LAYER 1: Base brief with 55+ techniques (already built, passed as baseBrief).
// LAYER 2: Corporate houses, negotiation timing, cash flow overlay, risk alerts.
// CROSS-LAYER: Convergences between base techniques and enterprise modules.
func BuildEnterpriseBrief(baseBrief string, chart *natal.Chart, year int) string {
	var b strings.Builder

	b.WriteString(baseBrief)
	b.WriteString("\n\n")
	b.WriteString("## ═══ CAPA EMPRESARIAL ═══\n\n")

	// Layer 2a: Corporate Houses
	corpHouses := CalcCorporateHouses(chart)
	if corpHouses != nil {
		b.WriteString("### CASAS CORPORATIVAS (Georgia Stathis)\n\n")
		for _, h := range corpHouses.Houses {
			line := fmt.Sprintf("- **Casa %d** (%s): %s", h.House, h.Theme, h.Interpretation)
			b.WriteString(line + "\n")
		}
		b.WriteString("\n")
	}

	// Layer 2b: Cash flow and risk from business module (direct calls, avoid full dashboard)
	cashFlow := CalcCashFlow(chart, year)
	if len(cashFlow) > 0 {
		b.WriteString("### CASH FLOW EMPRESARIAL\n\n")
		for _, cf := range cashFlow {
			if cf.Rating != "neutro" {
				b.WriteString(fmt.Sprintf("- **%s**: %s (neto: %.0f)\n", cf.Label, cf.Rating, cf.Net))
			}
		}
		b.WriteString("\n")
	}

	// Risk alerts (level >= 3)
	riskMap := CalcRiskHeatmap(chart, year)
	hasRisk := false
	for _, r := range riskMap {
		if r.Level >= 3 {
			if !hasRisk {
				b.WriteString("### ALERTAS DE RIESGO\n\n")
				hasRisk = true
			}
			b.WriteString(fmt.Sprintf("- **%s** mes %d: nivel %d — %s\n", r.Category, r.Month, r.Level, r.Alert))
		}
	}
	if hasRisk {
		b.WriteString("\n")
	}

	// Mercury Rx periods
	mercRx := CalcMercuryRx(year)
	if len(mercRx) > 0 {
		b.WriteString("### MERCURIO RETRÓGRADO\n\n")
		for _, mrx := range mercRx {
			b.WriteString(fmt.Sprintf("- Mes %d–%d: %s — %s\n", mrx.StartMonth, mrx.EndMonth, mrx.Sign, mrx.Impact))
		}
		b.WriteString("\n")
	}

	return b.String()
}
