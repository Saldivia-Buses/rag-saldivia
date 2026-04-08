package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Camionerou/rag-saldivia/services/astro/internal/intelligence"
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

// --- Proactive Alerts (Plan 13 Fase 13) ---

func (h *Handler) Alerts(w http.ResponseWriter, r *http.Request) {
	if h.q == nil {
		jsonError(w, "database not configured", http.StatusServiceUnavailable)
		return
	}
	tid, uid, err := tenantAndUser(r)
	if err != nil {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if n, err := strconv.Atoi(limitStr); err == nil && n > 0 && n <= 50 {
			limit = n
		}
	}

	// List contacts for this user
	contacts, err := h.q.ListContacts(r.Context(), repository.ListContactsParams{
		TenantID: tid, UserID: uid,
		Limit: int32(limit), Offset: 0,
	})
	if err != nil {
		jsonError(w, "failed to list contacts", http.StatusInternalServerError)
		return
	}

	// Convert to alert-scannable format
	var scannable []intelligence.ContactForAlert
	for _, c := range contacts {
		if !c.BirthDate.Valid {
			continue
		}
		bd := c.BirthDate.Time
		hour := 12.0 // default to noon if unknown
		if c.BirthTime.Valid {
			// pgtype.Time stores microseconds since midnight
			hour = float64(c.BirthTime.Microseconds) / 3_600_000_000.0
		}
		scannable = append(scannable, intelligence.ContactForAlert{
			ID:         fmt.Sprintf("%x", c.ID.Bytes),
			Name:       c.Name,
			BirthYear:  bd.Year(),
			BirthMonth: int(bd.Month()),
			BirthDay:   bd.Day(),
			BirthHour:  hour,
			Lat:        c.Lat,
			Lon:        c.Lon,
			Alt:        0,
			UTCOffset:  int(c.UtcOffset),
		})
	}

	year := time.Now().Year()
	alerts := intelligence.ScanAlerts(scannable, year, limit)
	jsonOK(w, map[string]interface{}{
		"alerts": alerts,
		"scanned": len(scannable),
	})
}
