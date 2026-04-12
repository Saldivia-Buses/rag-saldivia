package service

import (
	"time"

	"github.com/Camionerou/rag-saldivia/services/erp/internal/repository"
)

// Analytics handles reporting and BI queries. Read-only — no mutations, no audit log.
type Analytics struct {
	repo *repository.Queries
}

// NewAnalytics creates an analytics service.
func NewAnalytics(repo *repository.Queries) *Analytics {
	return &Analytics{repo: repo}
}

// Repo returns the repository for direct query access from the handler.
// Analytics is read-only, so direct repo access is safe.
func (a *Analytics) Repo() *repository.Queries {
	return a.repo
}

// DateRange holds common date filter params.
type DateRange struct {
	From time.Time
	To   time.Time
}

// DefaultDateRange returns the last 12 months.
func DefaultDateRange() DateRange {
	now := time.Now()
	return DateRange{
		From: time.Date(now.Year()-1, now.Month(), 1, 0, 0, 0, 0, time.UTC),
		To:   now,
	}
}
