package handler

import (
	"net/http"

	"github.com/Camionerou/rag-saldivia/services/astro/internal/repository"
)

// --- Daily Usage ---

func (h *Handler) DailyUsage(w http.ResponseWriter, r *http.Request) {
	if h.q == nil {
		jsonError(w, "database not configured", http.StatusServiceUnavailable)
		return
	}
	tid, uid, err := tenantAndUser(r)
	if err != nil {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	usage, err := h.q.GetUsageToday(r.Context(), repository.GetUsageTodayParams{
		TenantID: tid, UserID: uid,
	})
	if err != nil {
		// No usage today is not an error — return zeros
		jsonOK(w, map[string]int{"queries": 0, "tokens_in": 0, "tokens_out": 0})
		return
	}
	jsonOK(w, usage)
}
