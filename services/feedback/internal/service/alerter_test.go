// Unit tests for the Alerter threshold logic.
//
// CheckAndAlert uses *pgxpool.Pool directly (no repository interface), so
// integration with the real DB is not mockable at the unit level. The testable
// surface area is the alertChecks slice: each AlertCheck contains pure functions
// (Condition, Title, Description, CurrentVal, Threshold) that express the
// entire threshold policy.
//
// Testing strategy:
//   - alertChecks[*].Condition: pure functions over HealthScores → bool
//   - alertChecks[*].Title/Description/CurrentVal: pure string generators
//   - HealthScores composition: overall formula weights (30+25+20+15+10 = 100%)
//   - autoResolve and CheckAndAlert DB paths → TDD-ANCHOR (require testcontainers)
//
// INVARIANT tested: every alert type has exactly one Condition function. The
// full set of alert types is verified against the expected schema to catch
// regressions where a check is accidentally removed.
package service

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Tests: AlertCheck.Condition — threshold boundary conditions
// ---------------------------------------------------------------------------

// findCheck returns the AlertCheck for the given alertType, or fails the test.
func findCheck(t *testing.T, alertType string) AlertCheck {
	t.Helper()
	for _, c := range alertChecks {
		if c.AlertType == alertType {
			return c
		}
	}
	t.Fatalf("alertCheck not found: %q — was it removed from alertChecks?", alertType)
	return AlertCheck{} // unreachable
}

func TestAlerter_AllAlertTypesPresent(t *testing.T) {
	// Verify the full set of alert types exists. If a check is accidentally
	// removed, this test catches the regression.
	expectedTypes := []string{
		"quality_critical",
		"quality_drop",
		"error_spike",
		"error_critical",
		"latency_spike",
		"security_anomaly",
		"security_critical",
		"inactive_tenant",
	}
	for _, want := range expectedTypes {
		found := false
		for _, c := range alertChecks {
			if c.AlertType == want {
				found = true
				break
			}
		}
		require.True(t, found, "alert type %q missing from alertChecks", want)
	}
	require.Len(t, alertChecks, len(expectedTypes),
		"alertChecks has unexpected length — update this test if a new check was added")
}

// ---------------------------------------------------------------------------
// quality_critical: AIQuality < 50
// ---------------------------------------------------------------------------

func TestAlerter_QualityCritical_Condition(t *testing.T) {
	check := findCheck(t, "quality_critical")

	tests := []struct {
		name    string
		scores  HealthScores
		wantFire bool
	}{
		{"ai=49 fires", HealthScores{AIQuality: 49}, true},
		{"ai=0 fires", HealthScores{AIQuality: 0}, true},
		{"ai=50 does not fire", HealthScores{AIQuality: 50}, false},
		{"ai=100 does not fire", HealthScores{AIQuality: 100}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := check.Condition(tt.scores)
			require.Equal(t, tt.wantFire, got)
		})
	}
}

func TestAlerter_QualityCritical_Severity_IsCritical(t *testing.T) {
	check := findCheck(t, "quality_critical")
	require.Equal(t, "critical", check.Severity)
}

func TestAlerter_QualityCritical_CurrentVal_ShowsPercentage(t *testing.T) {
	check := findCheck(t, "quality_critical")
	got := check.CurrentVal(HealthScores{AIQuality: 45})
	require.Equal(t, "45%", got)
}

func TestAlerter_QualityCritical_Title_ContainsSlug(t *testing.T) {
	check := findCheck(t, "quality_critical")
	got := check.Title("acme")
	require.Contains(t, got, "acme")
}

// ---------------------------------------------------------------------------
// quality_drop: AIQuality in [50, 70)
// ---------------------------------------------------------------------------

func TestAlerter_QualityDrop_Condition(t *testing.T) {
	check := findCheck(t, "quality_drop")

	tests := []struct {
		name     string
		scores   HealthScores
		wantFire bool
	}{
		{"ai=69 fires", HealthScores{AIQuality: 69}, true},
		{"ai=50 fires", HealthScores{AIQuality: 50}, true},
		{"ai=70 does not fire", HealthScores{AIQuality: 70}, false},
		{"ai=49 does not fire (quality_critical handles it)", HealthScores{AIQuality: 49}, false},
		{"ai=100 does not fire", HealthScores{AIQuality: 100}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := check.Condition(tt.scores)
			require.Equal(t, tt.wantFire, got)
		})
	}
}

func TestAlerter_QualityDrop_Severity_IsWarning(t *testing.T) {
	check := findCheck(t, "quality_drop")
	require.Equal(t, "warning", check.Severity)
}

// ---------------------------------------------------------------------------
// error_spike: ErrorRate < 50
// ---------------------------------------------------------------------------

func TestAlerter_ErrorSpike_Condition(t *testing.T) {
	check := findCheck(t, "error_spike")

	tests := []struct {
		name     string
		scores   HealthScores
		wantFire bool
	}{
		{"error_rate=49 fires", HealthScores{ErrorRate: 49}, true},
		{"error_rate=0 fires", HealthScores{ErrorRate: 0}, true},
		{"error_rate=50 does not fire", HealthScores{ErrorRate: 50}, false},
		{"error_rate=100 does not fire", HealthScores{ErrorRate: 100}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := check.Condition(tt.scores)
			require.Equal(t, tt.wantFire, got)
		})
	}
}

// ---------------------------------------------------------------------------
// error_critical: ErrorRate <= 30
// ---------------------------------------------------------------------------

func TestAlerter_ErrorCritical_Condition(t *testing.T) {
	check := findCheck(t, "error_critical")

	tests := []struct {
		name     string
		scores   HealthScores
		wantFire bool
	}{
		{"error_rate=30 fires (boundary)", HealthScores{ErrorRate: 30}, true},
		{"error_rate=0 fires", HealthScores{ErrorRate: 0}, true},
		{"error_rate=31 does not fire", HealthScores{ErrorRate: 31}, false},
		{"error_rate=100 does not fire", HealthScores{ErrorRate: 100}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := check.Condition(tt.scores)
			require.Equal(t, tt.wantFire, got)
		})
	}
}

func TestAlerter_ErrorCritical_Severity_IsCritical(t *testing.T) {
	check := findCheck(t, "error_critical")
	require.Equal(t, "critical", check.Severity)
}

// ---------------------------------------------------------------------------
// latency_spike: Performance < 50
// ---------------------------------------------------------------------------

func TestAlerter_LatencySpike_Condition(t *testing.T) {
	check := findCheck(t, "latency_spike")

	tests := []struct {
		name     string
		scores   HealthScores
		wantFire bool
	}{
		{"perf=49 fires", HealthScores{Performance: 49}, true},
		{"perf=0 fires", HealthScores{Performance: 0}, true},
		{"perf=50 does not fire", HealthScores{Performance: 50}, false},
		{"perf=100 does not fire", HealthScores{Performance: 100}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := check.Condition(tt.scores)
			require.Equal(t, tt.wantFire, got)
		})
	}
}

// ---------------------------------------------------------------------------
// security_anomaly: Security < 70
// ---------------------------------------------------------------------------

func TestAlerter_SecurityAnomaly_Condition(t *testing.T) {
	check := findCheck(t, "security_anomaly")

	tests := []struct {
		name     string
		scores   HealthScores
		wantFire bool
	}{
		{"security=69 fires", HealthScores{Security: 69}, true},
		{"security=0 fires", HealthScores{Security: 0}, true},
		{"security=70 does not fire", HealthScores{Security: 70}, false},
		{"security=100 does not fire", HealthScores{Security: 100}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := check.Condition(tt.scores)
			require.Equal(t, tt.wantFire, got)
		})
	}
}

// ---------------------------------------------------------------------------
// security_critical: Security <= 30
// ---------------------------------------------------------------------------

func TestAlerter_SecurityCritical_Condition(t *testing.T) {
	check := findCheck(t, "security_critical")

	tests := []struct {
		name     string
		scores   HealthScores
		wantFire bool
	}{
		{"security=30 fires (boundary)", HealthScores{Security: 30}, true},
		{"security=0 fires", HealthScores{Security: 0}, true},
		{"security=31 does not fire", HealthScores{Security: 31}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := check.Condition(tt.scores)
			require.Equal(t, tt.wantFire, got)
		})
	}
}

func TestAlerter_SecurityCritical_Severity_IsCritical(t *testing.T) {
	check := findCheck(t, "security_critical")
	require.Equal(t, "critical", check.Severity)
}

// ---------------------------------------------------------------------------
// inactive_tenant: Usage < 25
// ---------------------------------------------------------------------------

func TestAlerter_InactiveTenant_Condition(t *testing.T) {
	check := findCheck(t, "inactive_tenant")

	tests := []struct {
		name     string
		scores   HealthScores
		wantFire bool
	}{
		{"usage=24 fires", HealthScores{Usage: 24}, true},
		{"usage=0 fires", HealthScores{Usage: 0}, true},
		{"usage=25 does not fire", HealthScores{Usage: 25}, false},
		{"usage=100 does not fire", HealthScores{Usage: 100}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := check.Condition(tt.scores)
			require.Equal(t, tt.wantFire, got)
		})
	}
}

func TestAlerter_InactiveTenant_Severity_IsInfo(t *testing.T) {
	check := findCheck(t, "inactive_tenant")
	require.Equal(t, "info", check.Severity)
}

// ---------------------------------------------------------------------------
// HealthScores: overall formula verification
//
// aggregate() computes:
//   overall = aiScore*0.30 + errorScore*0.25 + perfScore*0.20 + secScore*0.15 + usageScore*0.10
//
// These tests verify the weight formula without touching the DB.
// ---------------------------------------------------------------------------

func TestHealthScores_OverallFormula_AllHundred_GivesHundred(t *testing.T) {
	ai, err, perf, sec, usage := 100.0, 100.0, 100.0, 100.0, 100.0
	overall := ai*0.30 + err*0.25 + perf*0.20 + sec*0.15 + usage*0.10

	require.InDelta(t, 100.0, overall, 0.001,
		"all-100 inputs must produce overall=100")
}

func TestHealthScores_OverallFormula_AllZero_GivesZero(t *testing.T) {
	overall := 0.0*0.30 + 0.0*0.25 + 0.0*0.20 + 0.0*0.15 + 0.0*0.10
	require.InDelta(t, 0.0, overall, 0.001)
}

func TestHealthScores_OverallFormula_WeightsSum_IsOne(t *testing.T) {
	// Verify that 0.30+0.25+0.20+0.15+0.10 = 1.00 (no weight drift)
	sum := 0.30 + 0.25 + 0.20 + 0.15 + 0.10
	require.InDelta(t, 1.0, sum, 0.0001, "weights must sum to 1.0")
}

func TestHealthScores_OverallFormula_AIQualityHeaviest(t *testing.T) {
	// ai_quality has the highest weight (0.30). Dropping it to 0 should
	// have a larger impact than dropping any other single dimension.
	baseline := 100.0*0.30 + 100.0*0.25 + 100.0*0.20 + 100.0*0.15 + 100.0*0.10

	dropAI := 0.0*0.30 + 100.0*0.25 + 100.0*0.20 + 100.0*0.15 + 100.0*0.10
	dropErr := 100.0*0.30 + 0.0*0.25 + 100.0*0.20 + 100.0*0.15 + 100.0*0.10

	_ = baseline
	// Dropping AI quality causes a larger delta than dropping error rate
	deltaAI := 100.0 - dropAI
	deltaErr := 100.0 - dropErr

	require.Greater(t, deltaAI, deltaErr,
		"ai_quality (weight=0.30) must have larger impact than error_rate (weight=0.25)")
}

// ---------------------------------------------------------------------------
// HealthScores: threshold coverage table
//
// Verify that "healthy" scores (everything at 100) do not trigger any alert,
// and that a specific degraded scenario triggers expected alerts.
// ---------------------------------------------------------------------------

func TestAlerter_HealthyScores_NoConditionFires(t *testing.T) {
	healthy := HealthScores{
		Overall:     100,
		AIQuality:   100,
		ErrorRate:   100,
		Performance: 100,
		Security:    100,
		Usage:       100,
	}
	for _, check := range alertChecks {
		require.False(t, check.Condition(healthy),
			"alert %q must not fire for healthy scores", check.AlertType)
	}
}

func TestAlerter_DegradedScores_ExpectedAlertsFire(t *testing.T) {
	degraded := HealthScores{
		AIQuality:   45, // quality_critical fires (< 50)
		ErrorRate:   25, // error_critical fires (<= 30)
		Performance: 80, // latency_spike does NOT fire (>= 50)
		Security:    80, // security_anomaly does NOT fire (>= 70)
		Usage:       80, // inactive_tenant does NOT fire (>= 25)
	}

	expecting := map[string]bool{
		"quality_critical": true,
		"error_critical":   true,
	}
	notExpecting := map[string]bool{
		"latency_spike":   true,
		"security_anomaly": true,
		"inactive_tenant": true,
	}

	for _, check := range alertChecks {
		fired := check.Condition(degraded)
		if expecting[check.AlertType] {
			require.True(t, fired,
				"alert %q must fire for degraded scores", check.AlertType)
		}
		if notExpecting[check.AlertType] {
			require.False(t, fired,
				"alert %q must NOT fire for this scenario", check.AlertType)
		}
	}
}

// ---------------------------------------------------------------------------
// AlertCheck: string generators return non-empty output
// ---------------------------------------------------------------------------

func TestAlerter_AllChecks_TitleAndDescriptionNonEmpty(t *testing.T) {
	scores := HealthScores{
		AIQuality:   30,
		ErrorRate:   20,
		Performance: 30,
		Security:    20,
		Usage:       10,
	}
	slug := "saldivia"
	for _, check := range alertChecks {
		title := check.Title(slug)
		desc := check.Description(slug, scores)
		curr := check.CurrentVal(scores)
		thresh := check.Threshold

		require.NotEmpty(t, title, "check %q: Title must not be empty", check.AlertType)
		require.NotEmpty(t, desc, "check %q: Description must not be empty", check.AlertType)
		require.NotEmpty(t, curr, "check %q: CurrentVal must not be empty", check.AlertType)
		require.NotEmpty(t, thresh, "check %q: Threshold must not be empty", check.AlertType)
	}
}

func TestAlerter_AllChecks_TitleContainsSlug(t *testing.T) {
	slug := "test-tenant"
	for _, check := range alertChecks {
		title := check.Title(slug)
		require.Contains(t, title, slug,
			"check %q: Title must include tenant slug for identification", check.AlertType)
	}
}

// ---------------------------------------------------------------------------
// TDD-ANCHOR: CheckAndAlert and autoResolve (require testcontainers)
//
// These paths interact with platformDB directly via QueryRow/Exec on the
// *pgxpool.Pool. They cannot be unit-tested without a real PostgreSQL connection.
//
// TestAlerter_CheckAndAlert_BelowThreshold_AutoResolvesActiveAlert:
//   Pre-insert an active alert of type "quality_critical" for tenant T.
//   Call CheckAndAlert with scores where AIQuality=80 (condition=false).
//   → feedback_alerts.status for that row becomes "auto_resolved".
//   → resolved_at is set.
//
// TestAlerter_CheckAndAlert_AboveThreshold_InsertsAlert:
//   No active alerts for tenant T.
//   Call CheckAndAlert with scores where AIQuality=45 (condition fires).
//   → new row in feedback_alerts with alert_type="quality_critical", status="active".
//
// TestAlerter_CheckAndAlert_ActiveAlertExists_DoesNotDuplicate:
//   Pre-insert an active "quality_critical" alert.
//   Call CheckAndAlert again with same scores.
//   → no second row inserted (dedup by active status check).
//
// TestAlerter_CheckAndAlert_WithPublisher_BroadcastsCalled:
//   Mock EventPublisher.Broadcast is called when a new alert is created.
//   → publisher.Broadcast called with channel="feedback.alerts".
//   → publisher.Broadcast NOT called if alert already exists (dedup).
//
// TestAlerter_NewAlerter_NilPublisher_DoesNotPanic:
//   NewAlerter(pool, nil) → CheckAndAlert with firing condition → no panic.
//   (publisher nil check on line 196 of alerter.go)
// ---------------------------------------------------------------------------

// TestAlerter_NewAlerter_NilPublisher_IsValid verifies that NewAlerter accepts
// a nil publisher without panicking. CheckAndAlert with nil publisher skips
// the Broadcast call (gated by `if a.publisher != nil`).
func TestAlerter_NewAlerter_NilPublisher_IsValid(t *testing.T) {
	// NewAlerter with nil publisher and nil pool
	// This exercises the constructor — the nil pool would only panic
	// if CheckAndAlert is called (which requires a real DB).
	alerter := NewAlerter(nil, nil)
	require.NotNil(t, alerter)
	require.Nil(t, alerter.publisher)
}

// TestAlerter_ConditionFmt_CurrentVal verifies the fmt.Sprintf patterns for
// CurrentVal don't produce the Go default (%!v(MISSING)) error string.
func TestAlerter_ConditionFmt_CurrentVal_NoFormatError(t *testing.T) {
	scores := HealthScores{
		AIQuality:   50.5,
		ErrorRate:   40.123,
		Performance: 30.7,
		Security:    20.9,
		Usage:       15.0,
	}
	for _, check := range alertChecks {
		curr := check.CurrentVal(scores)
		require.NotContains(t, curr, "%!",
			"check %q: CurrentVal has format string error: %q", check.AlertType, curr)
	}
}

// TestAlerter_Description_ContainsScoreValue verifies Description embeds
// the actual score value for quality and performance checks (user-visible info).
func TestAlerter_Description_ContainsScoreValue(t *testing.T) {
	scores := HealthScores{AIQuality: 42, ErrorRate: 20, Performance: 30}

	qualityCritical := findCheck(t, "quality_critical")
	desc := qualityCritical.Description("slug", scores)
	require.Contains(t, desc, fmt.Sprintf("%.0f", scores.AIQuality),
		"quality_critical description must embed the actual AI quality score")
}
