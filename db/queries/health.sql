-- name: Health :one
SELECT 1;

-- name: GetSlaPolicy :one
SELECT priority, duration_seconds, updated_at FROM sla_policies WHERE priority = $1;
