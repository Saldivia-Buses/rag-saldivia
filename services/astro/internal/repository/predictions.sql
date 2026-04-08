-- name: CreatePrediction :one
INSERT INTO astro_predictions (tenant_id, user_id, session_id, contact_id, category, description, date_from, date_to, techniques)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, tenant_id, user_id, session_id, contact_id, category, description, date_from, date_to, techniques, outcome, outcome_notes, verified_at, created_at;

-- name: ListPredictions :many
SELECT id, tenant_id, user_id, session_id, contact_id, category, description, date_from, date_to, techniques, outcome, outcome_notes, verified_at, created_at
FROM astro_predictions WHERE tenant_id = $1 AND user_id = $2
ORDER BY created_at DESC LIMIT $3 OFFSET $4;

-- name: ListPendingPredictions :many
SELECT id, tenant_id, user_id, session_id, contact_id, category, description, date_from, date_to, techniques, outcome, outcome_notes, verified_at, created_at
FROM astro_predictions WHERE tenant_id = $1 AND user_id = $2 AND outcome = 'pending'
ORDER BY date_from ASC LIMIT $3;

-- name: VerifyPrediction :one
UPDATE astro_predictions SET outcome = $4, outcome_notes = $5, verified_at = now()
WHERE tenant_id = $1 AND user_id = $2 AND id = $3
RETURNING id, tenant_id, user_id, session_id, contact_id, category, description, date_from, date_to, techniques, outcome, outcome_notes, verified_at, created_at;

-- name: PredictionStats :one
SELECT
    count(*) AS total,
    count(*) FILTER (WHERE outcome = 'correct') AS correct,
    count(*) FILTER (WHERE outcome = 'incorrect') AS incorrect,
    count(*) FILTER (WHERE outcome = 'partial') AS partial,
    count(*) FILTER (WHERE outcome = 'pending') AS pending
FROM astro_predictions WHERE tenant_id = $1 AND user_id = $2;

-- name: CreateFeedback :one
INSERT INTO astro_feedback (tenant_id, message_id, user_id, thumbs, comment)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (message_id, user_id) DO UPDATE SET thumbs = $4, comment = $5
RETURNING id, tenant_id, message_id, user_id, thumbs, comment, created_at;

-- name: UpsertUsage :exec
INSERT INTO astro_usage (tenant_id, user_id, date, queries, tokens_in, tokens_out)
VALUES ($1, $2, CURRENT_DATE, $3, $4, $5)
ON CONFLICT (tenant_id, user_id, date)
DO UPDATE SET queries = astro_usage.queries + $3,
              tokens_in = astro_usage.tokens_in + $4,
              tokens_out = astro_usage.tokens_out + $5;

-- name: GetUsageToday :one
SELECT id, tenant_id, user_id, date, queries, tokens_in, tokens_out
FROM astro_usage WHERE tenant_id = $1 AND user_id = $2 AND date = CURRENT_DATE;
