package service

import (
	"context"
	"log/slog"
	"math"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Camionerou/rag-saldivia/services/feedback/internal/repository"
)

// Aggregator runs periodically to compute metrics and health scores.
type Aggregator struct {
	tenantDB    *pgxpool.Pool
	platformDB  *pgxpool.Pool
	repo        *repository.Queries
	feedbackSvc *Feedback
	alerter     *Alerter
	interval    time.Duration
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewAggregator creates an aggregator that runs on the given interval.
func NewAggregator(tenantDB, platformDB *pgxpool.Pool, feedbackSvc *Feedback, alerter *Alerter, interval time.Duration) *Aggregator {
	return &Aggregator{
		tenantDB:    tenantDB,
		platformDB:  platformDB,
		repo:        feedbackSvc.Repo(),
		feedbackSvc: feedbackSvc,
		alerter:     alerter,
		interval:    interval,
	}
}

// Start begins the aggregation loop.
func (a *Aggregator) Start(ctx context.Context, tenantID, tenantSlug string) {
	a.ctx, a.cancel = context.WithCancel(ctx)

	go func() {
		// Run once immediately on startup
		a.aggregate(tenantID, tenantSlug)

		ticker := time.NewTicker(a.interval)
		defer ticker.Stop()

		for {
			select {
			case <-a.ctx.Done():
				return
			case <-ticker.C:
				a.aggregate(tenantID, tenantSlug)
			}
		}
	}()

	slog.Info("feedback aggregator started", "interval", a.interval)
}

// Stop cancels the aggregation loop.
func (a *Aggregator) Stop() {
	if a.cancel != nil {
		a.cancel()
	}
}

func (a *Aggregator) aggregate(tenantID, tenantSlug string) {
	ctx := context.Background()
	period := time.Now().Truncate(time.Hour)

	slog.Debug("running feedback aggregation", "period", period, "tenant", tenantID)

	// Compute each dimension
	aiScore := a.computeAIQualityScore(ctx)
	errorScore := a.computeErrorRateScore(ctx)
	perfScore := a.computePerformanceScore(ctx)
	securityScore := a.computeSecurityScore(ctx)
	usageScore := a.computeUsageScore(ctx)

	// Weighted composite
	overall := aiScore*0.30 + errorScore*0.25 + perfScore*0.20 + securityScore*0.15 + usageScore*0.10

	// Upsert health score in platform DB (platform DB queries stay inline)
	_, err := a.platformDB.Exec(ctx,
		`INSERT INTO tenant_health_scores (tenant_id, period, overall_score, ai_quality_score, error_rate_score, usage_score, performance_score, security_score)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 ON CONFLICT (tenant_id, period) DO UPDATE SET
		   overall_score = EXCLUDED.overall_score,
		   ai_quality_score = EXCLUDED.ai_quality_score,
		   error_rate_score = EXCLUDED.error_rate_score,
		   usage_score = EXCLUDED.usage_score,
		   performance_score = EXCLUDED.performance_score,
		   security_score = EXCLUDED.security_score`,
		tenantID, period, overall, aiScore, errorScore, usageScore, perfScore, securityScore,
	)
	if err != nil {
		slog.Error("failed to upsert health score", "error", err, "tenant", tenantID)
		return
	}

	// Upsert aggregated metrics per category
	a.aggregateMetrics(ctx, tenantID, period)

	// Run alerter — check thresholds and create alerts
	if a.alerter != nil {
		scores := HealthScores{
			Overall:     overall,
			AIQuality:   aiScore,
			ErrorRate:   errorScore,
			Performance: perfScore,
			Security:    securityScore,
			Usage:       usageScore,
		}
		a.alerter.CheckAndAlert(ctx, tenantID, tenantSlug, scores)
	}

	// Purge old granular data (90 days)
	a.purgeOldEvents(ctx)

	slog.Info("feedback aggregation complete",
		"tenant", tenantID,
		"overall_score", math.Round(overall*10)/10,
		"ai", math.Round(aiScore*10)/10,
		"errors", math.Round(errorScore*10)/10,
		"perf", math.Round(perfScore*10)/10,
		"security", math.Round(securityScore*10)/10,
		"usage", math.Round(usageScore*10)/10,
	)
}

func (a *Aggregator) computeAIQualityScore(ctx context.Context) float64 {
	positive, negative, total, _, err := a.feedbackSvc.QualityMetrics(ctx, 1)
	if err != nil || total == 0 {
		return 100 // no data = assume OK
	}

	rated := positive + negative
	if rated == 0 {
		return 100 // events exist but none have thumbs — assume OK
	}

	positiveRate := float64(positive) / float64(rated)

	// If very few rated samples, attenuate the change
	if rated < 5 {
		return 50 + positiveRate*50 // range 50-100 for low sample
	}

	return positiveRate * 100
}

func (a *Aggregator) computeErrorRateScore(ctx context.Context) float64 {
	total, critical, _, err := a.feedbackSvc.ErrorCounts(ctx, 1)
	if err != nil {
		return 100
	}

	if critical > 0 {
		return math.Max(0, math.Min(30, float64(100-total*5)))
	}

	switch {
	case total == 0:
		return 100
	case total <= 2:
		return 90
	case total <= 5:
		return 70
	case total <= 10:
		return 50
	case total <= 20:
		return 30
	default:
		return 10
	}
}

func (a *Aggregator) computePerformanceScore(ctx context.Context) float64 {
	_, p95, _, err := a.feedbackSvc.PerformancePercentiles(ctx, 1)
	if err != nil || p95 == 0 {
		return 100 // no data
	}

	switch {
	case p95 < 200:
		return 100
	case p95 < 500:
		return 85
	case p95 < 1000:
		return 70
	case p95 < 2000:
		return 50
	case p95 < 5000:
		return 30
	default:
		return 10
	}
}

func (a *Aggregator) computeSecurityScore(ctx context.Context) float64 {
	counts, err := a.feedbackSvc.CountByCategory(ctx, 1)
	if err != nil {
		return 100
	}

	securityEvents := counts["security"]
	if securityEvents == 0 {
		return 100
	}

	// Check for critical severity via repository
	critCount, err := a.repo.CountCriticalSecurityEvents(ctx)
	if err != nil {
		slog.Error("failed to count critical security events", "error", err)
		return 70 // degrade gracefully
	}

	if critCount > 1 {
		return 10
	}
	if critCount == 1 {
		return 30
	}

	switch {
	case securityEvents <= 3:
		return 85
	case securityEvents <= 10:
		return 70
	default:
		return 50
	}
}

func (a *Aggregator) computeUsageScore(ctx context.Context) float64 {
	counts, err := a.feedbackSvc.CountByCategory(ctx, 1)
	if err != nil {
		return 50 // neutral for new tenant
	}

	currentUsage := counts["usage"]
	if currentUsage == 0 {
		// Check if this is a new tenant (no historical data)
		historicalCount, err := a.repo.CountHistoricalUsage(ctx)
		if err != nil {
			slog.Error("failed to count historical usage", "error", err)
			return 50
		}

		if historicalCount == 0 {
			return 50 // new tenant, neutral
		}
		return 20 // has history but zero current usage
	}

	// Compare to 7-day average
	avgHourly, err := a.repo.AvgHourlyUsage(ctx)
	if err != nil {
		slog.Error("failed to get avg hourly usage", "error", err)
		return 100
	}

	if avgHourly == 0 {
		return 100
	}

	ratio := float64(currentUsage) / avgHourly
	switch {
	case ratio >= 0.5:
		return 100
	case ratio >= 0.2:
		return 70
	case ratio >= 0.05:
		return 40
	default:
		return 20
	}
}

func (a *Aggregator) aggregateMetrics(ctx context.Context, tenantID string, period time.Time) {
	rows, err := a.repo.AggregateByModuleCategory(ctx, repository.AggregateByModuleCategoryParams{
		CreatedAt:   pgtype.Timestamptz{Time: period.Add(-a.interval), Valid: true},
		CreatedAt_2: pgtype.Timestamptz{Time: period, Valid: true},
	})
	if err != nil {
		slog.Error("failed to query aggregate data", "error", err)
		return
	}

	for _, row := range rows {
		// Platform DB upsert stays inline (different database)
		_, err := a.platformDB.Exec(ctx,
			`INSERT INTO feedback_metrics (tenant_id, module, category, period, total_events, positive, negative, avg_score, error_count)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			 ON CONFLICT (tenant_id, module, category, period) DO UPDATE SET
			   total_events = EXCLUDED.total_events,
			   positive = EXCLUDED.positive,
			   negative = EXCLUDED.negative,
			   avg_score = EXCLUDED.avg_score,
			   error_count = EXCLUDED.error_count`,
			tenantID, row.Module, row.Category, period,
			row.Total, row.Positive, row.Negative, row.AvgScore, row.ErrorCount,
		)
		if err != nil {
			slog.Error("failed to upsert metric", "error", err, "module", row.Module, "category", row.Category)
		}
	}
}

func (a *Aggregator) purgeOldEvents(ctx context.Context) {
	count, err := a.repo.PurgeOldEvents(ctx)
	if err != nil {
		slog.Error("failed to purge old feedback events", "error", err)
		return
	}
	if count > 0 {
		slog.Info("purged old feedback events", "count", count)
	}
}
