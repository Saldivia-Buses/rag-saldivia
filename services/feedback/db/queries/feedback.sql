-- name: InsertFeedbackEvent :exec
INSERT INTO feedback_events (category, module, user_id, score, thumbs, severity, context, comment)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: CountByCategory :many
SELECT category, COUNT(*)::int AS count
FROM feedback_events
WHERE created_at > now() - make_interval(hours => $1)
GROUP BY category;

-- name: QualityMetrics :one
SELECT
    COUNT(*) FILTER (WHERE thumbs = 'up')::int AS positive,
    COUNT(*) FILTER (WHERE thumbs = 'down')::int AS negative,
    COUNT(*)::int AS total,
    COALESCE(AVG(score), 0)::float8 AS avg_score
FROM feedback_events
WHERE category IN ('response_quality', 'agent_quality', 'extraction', 'detection')
  AND created_at > now() - make_interval(hours => $1);

-- name: ErrorCounts :one
SELECT
    COUNT(*)::int AS total,
    COUNT(*) FILTER (WHERE severity = 'critical')::int AS critical,
    COUNT(*) FILTER (WHERE status = 'open')::int AS open
FROM feedback_events
WHERE category = 'error_report'
  AND created_at > now() - make_interval(hours => $1);

-- name: PerformancePercentiles :one
SELECT
    COALESCE(percentile_cont(0.50) WITHIN GROUP (ORDER BY (context->>'latency_ms')::numeric), 0)::float8 AS p50,
    COALESCE(percentile_cont(0.95) WITHIN GROUP (ORDER BY (context->>'latency_ms')::numeric), 0)::float8 AS p95,
    COALESCE(percentile_cont(0.99) WITHIN GROUP (ORDER BY (context->>'latency_ms')::numeric), 0)::float8 AS p99
FROM feedback_events
WHERE category = 'performance'
  AND created_at > now() - make_interval(hours => $1);

-- name: CountCriticalSecurityEvents :one
SELECT COUNT(*)::int AS count
FROM feedback_events
WHERE category = 'security' AND severity = 'critical'
  AND created_at > now() - interval '1 hour';

-- name: CountHistoricalUsage :one
SELECT COUNT(*)::int AS count
FROM feedback_events
WHERE category = 'usage' AND created_at > $1;

-- name: AvgHourlyUsage :one
SELECT COALESCE(COUNT(*)::float8 / GREATEST(EXTRACT(EPOCH FROM (now() - MIN(created_at))) / 3600, 1), 0)::float8 AS avg_hourly
FROM feedback_events
WHERE category = 'usage'
  AND created_at > now() - interval '7 days';

-- name: AggregateByModuleCategory :many
SELECT module, category,
    COUNT(*)::int AS total,
    COUNT(*) FILTER (WHERE thumbs = 'up')::int AS positive,
    COUNT(*) FILTER (WHERE thumbs = 'down')::int AS negative,
    AVG(score)::float8 AS avg_score,
    COUNT(*) FILTER (WHERE category = 'error_report')::int AS error_count
FROM feedback_events
WHERE created_at > $1 AND created_at <= $2
GROUP BY module, category;

-- name: PurgeOldEvents :execrows
DELETE FROM feedback_events WHERE created_at < now() - interval '90 days';

-- name: GetSummaryAIQuality :one
SELECT COUNT(*)::int AS total_feedback,
    COALESCE(AVG(score), 0)::float8 AS avg_score,
    CASE WHEN COUNT(*) > 0 THEN COUNT(*) FILTER (WHERE thumbs = 'up')::float8 / NULLIF(COUNT(*), 0)
    ELSE 0 END::float8 AS positive_rate
FROM feedback_events
WHERE category IN ('response_quality','agent_quality','extraction','detection')
  AND created_at > now() - make_interval(hours => $1);

-- name: GetSummaryErrors :one
SELECT COUNT(*)::int AS total,
    COUNT(*) FILTER (WHERE status = 'open')::int AS open,
    COUNT(*) FILTER (WHERE severity = 'critical')::int AS critical
FROM feedback_events
WHERE category = 'error_report'
  AND created_at > now() - make_interval(hours => $1);

-- name: GetSummaryFeatures :one
SELECT COUNT(*)::int AS total,
    COUNT(*) FILTER (WHERE status = 'open')::int AS open
FROM feedback_events
WHERE category = 'feature_request'
  AND created_at > now() - make_interval(hours => $1);

-- name: GetSummaryNPS :one
SELECT COUNT(*)::int AS responses,
    COALESCE(
        (COUNT(*) FILTER (WHERE score >= 9)::float8 - COUNT(*) FILTER (WHERE score < 7)::float8)
        / NULLIF(COUNT(*), 0) * 100, 0)::float8 AS score
FROM feedback_events
WHERE category = 'nps'
  AND created_at > now() - interval '30 days';

-- name: ListQualityEvents :many
SELECT id, category, module, score, thumbs, comment, created_at
FROM feedback_events
WHERE category IN ('response_quality','agent_quality','extraction','detection')
  AND created_at > now() - make_interval(hours => $1)
ORDER BY created_at DESC
LIMIT $2;

-- name: ListQualityEventsByModule :many
SELECT id, category, module, score, thumbs, comment, created_at
FROM feedback_events
WHERE category IN ('response_quality','agent_quality','extraction','detection')
  AND created_at > now() - make_interval(hours => $1)
  AND module = $2
ORDER BY created_at DESC
LIMIT $3;

-- name: ListErrorEvents :many
SELECT id, module, severity, status, context, comment, created_at
FROM feedback_events
WHERE category = 'error_report' AND status = $1
ORDER BY created_at DESC
LIMIT $2;

-- name: GetUsageByModule :many
SELECT module, COUNT(*)::int AS actions, COUNT(DISTINCT user_id)::int AS unique_users
FROM feedback_events
WHERE category = 'usage'
  AND created_at > now() - make_interval(hours => $1)
GROUP BY module
ORDER BY COUNT(*) DESC;
