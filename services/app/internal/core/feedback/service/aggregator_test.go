// Unit tests for the Aggregator scoring functions.
//
// Architecture: computeAIQualityScore, computeErrorRateScore, and
// computePerformanceScore call feedbackSvc methods that use repository.DBTX
// (mockable via mockFeedbackDB from feedback_test.go). We can drive those paths
// through a *Aggregator constructed with a *Feedback backed by mockFeedbackDB.
//
// computeSecurityScore and computeUsageScore call feedbackSvc.CountByCategory
// which uses pgx.Rows (Query — mock returns nil, causes panic). Those paths
// are documented as TDD-ANCHOR integration tests.
//
// The aggregate() orchestration and Start/Stop lifecycle are also documented
// as TDD-ANCHOR (they require both tenant and platform DB pools).
//
// What IS testable without a real DB:
//   - computeAIQualityScore formula via mockFeedbackDB.QueryRow
//   - computeErrorRateScore formula via mockFeedbackDB.QueryRow
//   - computePerformanceScore formula via mockFeedbackDB.QueryRow
//   - aggregator.Stop() when cancel is nil (must not panic)
//   - overall weight composition (pure arithmetic)
package service

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"

	"github.com/Camionerou/rag-saldivia/services/app/internal/core/feedback/repository"
)

// ---------------------------------------------------------------------------
// Mock: repository.DBTX for aggregator tests
//
// We reuse the same DBTX interface but need a separate mock that can
// serve multi-column QueryRow results for QualityMetrics, ErrorCounts, and
// PerformancePercentiles.
// ---------------------------------------------------------------------------

// aggMockDB is a DBTX mock that serves a sequence of mockAggRow instances.
// Each call to QueryRow pops the next row from the queue.
type aggMockDB struct {
	rows []*mockAggRow
	idx  int
}

func (m *aggMockDB) Exec(_ context.Context, _ string, _ ...interface{}) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}

func (m *aggMockDB) Query(_ context.Context, _ string, _ ...interface{}) (pgx.Rows, error) {
	// aggMockDB does not simulate Query (pgx.Rows is too complex to mock).
	// computeSecurityScore and computeUsageScore are TDD-ANCHOR for integration tests.
	return nil, errors.New("aggMockDB: Query not supported — use integration tests")
}

func (m *aggMockDB) QueryRow(_ context.Context, _ string, _ ...interface{}) pgx.Row {
	if m.idx < len(m.rows) {
		r := m.rows[m.idx]
		m.idx++
		return r
	}
	return &mockAggRow{scanFn: func(...any) error { return pgx.ErrNoRows }}
}

// mockAggRow implements pgx.Row with a custom scan function.
type mockAggRow struct {
	scanFn func(dest ...any) error
}

func (r *mockAggRow) Scan(dest ...any) error {
	if r.scanFn != nil {
		return r.scanFn(dest...)
	}
	return pgx.ErrNoRows
}

// ---------------------------------------------------------------------------
// Builder: Aggregator backed by aggMockDB
// ---------------------------------------------------------------------------

// newTestAggregator builds an Aggregator where feedbackSvc is backed by
// aggMockDB. platformDB and tenantDB are nil — they are only used in
// aggregate() (TDD-ANCHOR).
func newTestAggregator(db *aggMockDB) *Aggregator {
	feedbackSvc := &Feedback{
		repo: repository.New(db),
	}
	return &Aggregator{
		feedbackSvc: feedbackSvc,
		repo:        feedbackSvc.Repo(),
		ctx:         context.Background(),
	}
}

// ---------------------------------------------------------------------------
// Row builders for specific QualityMetrics / ErrorCounts / PerformancePercentiles
// ---------------------------------------------------------------------------

// qualityMetricsRow scans: positive int32, negative int32, total int32, avgScore float64
func qualityMetricsRow(positive, negative, total int32, avgScore float64) *mockAggRow {
	return &mockAggRow{
		scanFn: func(dest ...any) error {
			if len(dest) < 4 {
				return errors.New("unexpected column count for QualityMetrics")
			}
			*(dest[0].(*int32)) = positive
			*(dest[1].(*int32)) = negative
			*(dest[2].(*int32)) = total
			*(dest[3].(*float64)) = avgScore
			return nil
		},
	}
}

// errorCountsRow scans: total int32, critical int32, open int32
func errorCountsRow(total, critical, open int32) *mockAggRow {
	return &mockAggRow{
		scanFn: func(dest ...any) error {
			if len(dest) < 3 {
				return errors.New("unexpected column count for ErrorCounts")
			}
			*(dest[0].(*int32)) = total
			*(dest[1].(*int32)) = critical
			*(dest[2].(*int32)) = open
			return nil
		},
	}
}

// perfPercentilesRow scans: p50 float64, p95 float64, p99 float64
func perfPercentilesRow(p50, p95, p99 float64) *mockAggRow {
	return &mockAggRow{
		scanFn: func(dest ...any) error {
			if len(dest) < 3 {
				return errors.New("unexpected column count for PerformancePercentiles")
			}
			*(dest[0].(*float64)) = p50
			*(dest[1].(*float64)) = p95
			*(dest[2].(*float64)) = p99
			return nil
		},
	}
}

// ---------------------------------------------------------------------------
// Tests: computeAIQualityScore
//
// Formula (from aggregator.go):
//   - total=0 or error → 100 (no data = assume OK)
//   - rated (positive+negative)=0 → 100 (events but no thumbs)
//   - rated<5 → 50 + positiveRate*50 (attenuated)
//   - rated>=5 → positiveRate * 100
// ---------------------------------------------------------------------------

func TestAggregator_AIQualityScore_NoData_Returns100(t *testing.T) {
	db := &aggMockDB{
		rows: []*mockAggRow{
			qualityMetricsRow(0, 0, 0, 0), // total=0
		},
	}
	agg := newTestAggregator(db)
	got := agg.computeAIQualityScore(context.Background())
	require.InDelta(t, 100.0, got, 0.001, "no data must return 100 (assume OK)")
}

func TestAggregator_AIQualityScore_NoThumbs_Returns100(t *testing.T) {
	// total=5 but positive=0, negative=0 (events exist, nobody rated)
	db := &aggMockDB{
		rows: []*mockAggRow{
			qualityMetricsRow(0, 0, 5, 0),
		},
	}
	agg := newTestAggregator(db)
	got := agg.computeAIQualityScore(context.Background())
	require.InDelta(t, 100.0, got, 0.001, "no thumbs → assume OK → 100")
}

func TestAggregator_AIQualityScore_LowSample_Attenuated(t *testing.T) {
	// rated=3 (< 5): 2 positive, 1 negative → positiveRate=2/3≈0.667
	// formula: 50 + 0.667*50 ≈ 83.33
	db := &aggMockDB{
		rows: []*mockAggRow{
			qualityMetricsRow(2, 1, 3, 4.0),
		},
	}
	agg := newTestAggregator(db)
	got := agg.computeAIQualityScore(context.Background())
	expected := 50 + (2.0/3.0)*50
	require.InDelta(t, expected, got, 0.5, "low sample (<5) must use attenuated formula")
	// Must be in [50,100]
	require.GreaterOrEqual(t, got, 50.0)
	require.LessOrEqual(t, got, 100.0)
}

func TestAggregator_AIQualityScore_FullSample_AllPositive(t *testing.T) {
	// rated=10 (>=5): all positive → positiveRate=1.0 → score=100
	db := &aggMockDB{
		rows: []*mockAggRow{
			qualityMetricsRow(10, 0, 10, 5.0),
		},
	}
	agg := newTestAggregator(db)
	got := agg.computeAIQualityScore(context.Background())
	require.InDelta(t, 100.0, got, 0.001, "all-positive rated>=5 must return 100")
}

func TestAggregator_AIQualityScore_FullSample_HalfNegative(t *testing.T) {
	// rated=10: 5 positive, 5 negative → positiveRate=0.5 → score=50
	db := &aggMockDB{
		rows: []*mockAggRow{
			qualityMetricsRow(5, 5, 10, 3.0),
		},
	}
	agg := newTestAggregator(db)
	got := agg.computeAIQualityScore(context.Background())
	require.InDelta(t, 50.0, got, 0.001, "50%% positive rated>=5 must return 50")
}

func TestAggregator_AIQualityScore_FullSample_MostlyNegative(t *testing.T) {
	// rated=10: 2 positive, 8 negative → positiveRate=0.2 → score=20
	db := &aggMockDB{
		rows: []*mockAggRow{
			qualityMetricsRow(2, 8, 10, 2.0),
		},
	}
	agg := newTestAggregator(db)
	got := agg.computeAIQualityScore(context.Background())
	require.InDelta(t, 20.0, got, 0.001, "80%% negative rated>=5 must return 20")
}

func TestAggregator_AIQualityScore_DBError_Returns100(t *testing.T) {
	// QueryRow returns ErrNoRows → QualityMetrics returns error → assume OK
	db := &aggMockDB{} // empty rows → ErrNoRows
	agg := newTestAggregator(db)
	got := agg.computeAIQualityScore(context.Background())
	require.InDelta(t, 100.0, got, 0.001, "DB error must return 100 (graceful degradation)")
}

// ---------------------------------------------------------------------------
// Tests: computeErrorRateScore
//
// Formula (from aggregator.go):
//   - error (DB): 100
//   - critical>0: math.Max(0, math.Min(30, 100-total*5))
//   - total=0: 100
//   - total<=2: 90
//   - total<=5: 70
//   - total<=10: 50
//   - total<=20: 30
//   - default: 10
// ---------------------------------------------------------------------------

func TestAggregator_ErrorRateScore_NoErrors_Returns100(t *testing.T) {
	db := &aggMockDB{rows: []*mockAggRow{errorCountsRow(0, 0, 0)}}
	agg := newTestAggregator(db)
	got := agg.computeErrorRateScore(context.Background())
	require.InDelta(t, 100.0, got, 0.001)
}

func TestAggregator_ErrorRateScore_FewErrors_Returns90(t *testing.T) {
	db := &aggMockDB{rows: []*mockAggRow{errorCountsRow(2, 0, 1)}}
	agg := newTestAggregator(db)
	got := agg.computeErrorRateScore(context.Background())
	require.InDelta(t, 90.0, got, 0.001, "total<=2, no critical → 90")
}

func TestAggregator_ErrorRateScore_MediumErrors_Returns70(t *testing.T) {
	db := &aggMockDB{rows: []*mockAggRow{errorCountsRow(5, 0, 2)}}
	agg := newTestAggregator(db)
	got := agg.computeErrorRateScore(context.Background())
	require.InDelta(t, 70.0, got, 0.001, "total<=5, no critical → 70")
}

func TestAggregator_ErrorRateScore_HighErrors_Returns50(t *testing.T) {
	db := &aggMockDB{rows: []*mockAggRow{errorCountsRow(10, 0, 3)}}
	agg := newTestAggregator(db)
	got := agg.computeErrorRateScore(context.Background())
	require.InDelta(t, 50.0, got, 0.001, "total<=10, no critical → 50")
}

func TestAggregator_ErrorRateScore_VeryHighErrors_Returns30(t *testing.T) {
	db := &aggMockDB{rows: []*mockAggRow{errorCountsRow(20, 0, 5)}}
	agg := newTestAggregator(db)
	got := agg.computeErrorRateScore(context.Background())
	require.InDelta(t, 30.0, got, 0.001, "total<=20, no critical → 30")
}

func TestAggregator_ErrorRateScore_ExtremeErrors_Returns10(t *testing.T) {
	db := &aggMockDB{rows: []*mockAggRow{errorCountsRow(21, 0, 10)}}
	agg := newTestAggregator(db)
	got := agg.computeErrorRateScore(context.Background())
	require.InDelta(t, 10.0, got, 0.001, "total>20, no critical → 10")
}

func TestAggregator_ErrorRateScore_CriticalError_LimitedTo30(t *testing.T) {
	// critical>0: math.Max(0, math.Min(30, 100-1*5)) = math.Min(30,95) = 30
	db := &aggMockDB{rows: []*mockAggRow{errorCountsRow(1, 1, 1)}}
	agg := newTestAggregator(db)
	got := agg.computeErrorRateScore(context.Background())
	require.LessOrEqual(t, got, 30.0, "critical error must cap score at 30")
}

func TestAggregator_ErrorRateScore_ManyCriticals_BottomsAtZero(t *testing.T) {
	// critical>0, total=100: math.Max(0, math.Min(30, 100-100*5)) = math.Max(0,-400)=0
	db := &aggMockDB{rows: []*mockAggRow{errorCountsRow(100, 5, 50)}}
	agg := newTestAggregator(db)
	got := agg.computeErrorRateScore(context.Background())
	require.InDelta(t, 0.0, got, 0.001, "many errors + critical must bottom at 0")
}

func TestAggregator_ErrorRateScore_DBError_Returns100(t *testing.T) {
	db := &aggMockDB{} // ErrNoRows → error → return 100
	agg := newTestAggregator(db)
	got := agg.computeErrorRateScore(context.Background())
	require.InDelta(t, 100.0, got, 0.001, "DB error → graceful degradation → 100")
}

// ---------------------------------------------------------------------------
// Tests: computePerformanceScore
//
// Formula (from aggregator.go — uses p95):
//   - error or p95=0: 100
//   - p95<200: 100
//   - p95<500: 85
//   - p95<1000: 70
//   - p95<2000: 50
//   - p95<5000: 30
//   - default: 10
// ---------------------------------------------------------------------------

func TestAggregator_PerformanceScore_NoData_Returns100(t *testing.T) {
	db := &aggMockDB{rows: []*mockAggRow{perfPercentilesRow(0, 0, 0)}} // p95=0
	agg := newTestAggregator(db)
	got := agg.computePerformanceScore(context.Background())
	require.InDelta(t, 100.0, got, 0.001, "p95=0 (no data) → 100")
}

func TestAggregator_PerformanceScore_ExcellentLatency_Returns100(t *testing.T) {
	db := &aggMockDB{rows: []*mockAggRow{perfPercentilesRow(50, 150, 180)}} // p95=150
	agg := newTestAggregator(db)
	got := agg.computePerformanceScore(context.Background())
	require.InDelta(t, 100.0, got, 0.001, "p95<200ms → 100")
}

func TestAggregator_PerformanceScore_GoodLatency_Returns85(t *testing.T) {
	db := &aggMockDB{rows: []*mockAggRow{perfPercentilesRow(100, 350, 480)}} // p95=350
	agg := newTestAggregator(db)
	got := agg.computePerformanceScore(context.Background())
	require.InDelta(t, 85.0, got, 0.001, "p95 in [200,500) → 85")
}

func TestAggregator_PerformanceScore_AcceptableLatency_Returns70(t *testing.T) {
	db := &aggMockDB{rows: []*mockAggRow{perfPercentilesRow(200, 750, 950)}} // p95=750
	agg := newTestAggregator(db)
	got := agg.computePerformanceScore(context.Background())
	require.InDelta(t, 70.0, got, 0.001, "p95 in [500,1000) → 70")
}

func TestAggregator_PerformanceScore_SlowLatency_Returns50(t *testing.T) {
	db := &aggMockDB{rows: []*mockAggRow{perfPercentilesRow(500, 1500, 1900)}} // p95=1500
	agg := newTestAggregator(db)
	got := agg.computePerformanceScore(context.Background())
	require.InDelta(t, 50.0, got, 0.001, "p95 in [1000,2000) → 50")
}

func TestAggregator_PerformanceScore_VerySlowLatency_Returns30(t *testing.T) {
	db := &aggMockDB{rows: []*mockAggRow{perfPercentilesRow(1000, 3000, 4500)}} // p95=3000
	agg := newTestAggregator(db)
	got := agg.computePerformanceScore(context.Background())
	require.InDelta(t, 30.0, got, 0.001, "p95 in [2000,5000) → 30")
}

func TestAggregator_PerformanceScore_CriticalLatency_Returns10(t *testing.T) {
	db := &aggMockDB{rows: []*mockAggRow{perfPercentilesRow(2000, 6000, 8000)}} // p95=6000
	agg := newTestAggregator(db)
	got := agg.computePerformanceScore(context.Background())
	require.InDelta(t, 10.0, got, 0.001, "p95>=5000ms → 10")
}

func TestAggregator_PerformanceScore_DBError_Returns100(t *testing.T) {
	db := &aggMockDB{} // ErrNoRows → error → 100
	agg := newTestAggregator(db)
	got := agg.computePerformanceScore(context.Background())
	require.InDelta(t, 100.0, got, 0.001, "DB error → graceful degradation → 100")
}

// ---------------------------------------------------------------------------
// Tests: Stop (safe to call when not started)
// ---------------------------------------------------------------------------

func TestAggregator_Stop_WhenNotStarted_DoesNotPanic(t *testing.T) {
	agg := &Aggregator{cancel: nil}
	require.NotPanics(t, func() {
		agg.Stop()
	}, "Stop() with nil cancel must not panic")
}

func TestAggregator_Stop_AfterCancel_IsIdempotent(t *testing.T) {
	ctx := context.Background()
	agg := &Aggregator{}
	agg.ctx, agg.cancel = context.WithCancel(ctx)

	// Stop once
	require.NotPanics(t, func() { agg.Stop() })
	// Stop again (cancel already called) — must not panic
	require.NotPanics(t, func() { agg.Stop() })
}

// ---------------------------------------------------------------------------
// Tests: NewAggregator constructor
// ---------------------------------------------------------------------------

func TestAggregator_NewAggregator_NilAlerter_IsValid(t *testing.T) {
	// NewAggregator accepts nil alerter — aggregate() gates the alerter call
	// with `if a.alerter != nil`. Verify the constructor doesn't panic with nil.
	agg := NewAggregator(nil, nil, &Feedback{repo: repository.New(&aggMockDB{})}, nil, 0)
	require.NotNil(t, agg)
	require.Nil(t, agg.alerter)
}

// ---------------------------------------------------------------------------
// TDD-ANCHOR: paths requiring real DB
//
// computeSecurityScore — calls feedbackSvc.CountByCategory (uses pgx.Rows Query)
//   then repo.CountCriticalSecurityEvents (QueryRow). The Query path cannot
//   be simulated with aggMockDB. Integration test contract:
//     Insert security events (1 critical) → computeSecurityScore returns 30.
//     Insert security events (0 critical, 5 events) → returns 85.
//     No security events → returns 100.
//
// computeUsageScore — calls feedbackSvc.CountByCategory (Query path)
//   then repo.CountHistoricalUsage or repo.AvgHourlyUsage (QueryRow).
//   Integration test contract:
//     No usage events, no historical → returns 50 (new tenant neutral).
//     No usage events, has historical → returns 20 (inactive).
//     Usage/avg ratio >= 0.5 → returns 100.
//     Usage/avg ratio in [0.2,0.5) → returns 70.
//
// aggregate() — calls all compute* and upserts to platformDB.
//   Requires both tenant pool (for compute* via feedbackSvc) and platform pool
//   (for INSERT INTO tenant_health_scores).
//
// Start()/Stop() loop — requires real ticker; tested via time.AfterFunc or
//   integration harness that lets the goroutine fire at least once.
//
// aggregateMetrics() — calls repo.AggregateByModuleCategory (Query) then
//   platformDB.Exec per row. Requires both pools.
//
// purgeOldEvents() — calls repo.PurgeOldEvents (Exec on tenant pool).
//   Integration test: pre-insert events >90 days old → verify they are removed.
// ---------------------------------------------------------------------------
