-- name: CreateUser :exec
INSERT INTO users (id, email, password_hash, role, status)
VALUES ($1, $2, $3, $4, $5);

-- name: GetUserByID :one
SELECT id, email, password_hash, role, status, created_at
FROM users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT id, email, password_hash, role, status, created_at
FROM users WHERE email = $1;

-- name: CreateTicket :exec
INSERT INTO tickets (id, title, description, customer_id, status, priority, category, sla_due_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: GetTicketByID :one
SELECT id, title, description, customer_id, assignee_id, team_id, status, priority, category,
  sla_due_at, sla_paused_at, sla_remaining_seconds, breached_at, created_at, updated_at
FROM tickets WHERE id = $1;

-- name: AddComment :exec
INSERT INTO comments (id, ticket_id, author_id, body, visibility)
VALUES ($1, $2, $3, $4, $5);

-- name: ListPublicComments :many
SELECT id, ticket_id, author_id, body, visibility, created_at
FROM comments WHERE ticket_id = $1 AND deleted_at IS NULL AND visibility = 'public'
ORDER BY created_at;
