package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/XoDeR/customer-support-desk-go/internal/domain/entity"
	"github.com/XoDeR/customer-support-desk-go/internal/infrastructure/database"
	"github.com/XoDeR/customer-support-desk-go/pkg/uuidv7"
)

type Repository struct{ db *sql.DB }

func New(db *sql.DB) *Repository { return &Repository{db} }
func (r *Repository) q(ctx context.Context) interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
} {
	if tx, ok := database.Tx(ctx); ok {
		return tx
	}
	return r.db
}
func userScan(row interface{ Scan(...any) error }) (entity.User, error) {
	var u entity.User
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.Status, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return u, entity.ErrNotFound
	}
	return u, err
}
func (r *Repository) GetUserByEmail(ctx context.Context, email string) (entity.User, error) {
	return userScan(r.q(ctx).QueryRowContext(ctx, `SELECT id,email,password_hash,role,status,created_at FROM users WHERE email=$1`, strings.ToLower(email)))
}
func (r *Repository) GetUser(ctx context.Context, id uuidv7.UUID) (entity.User, error) {
	return userScan(r.q(ctx).QueryRowContext(ctx, `SELECT id,email,password_hash,role,status,created_at FROM users WHERE id=$1`, id))
}
func (r *Repository) CreateUser(ctx context.Context, u entity.User) error {
	_, err := r.q(ctx).ExecContext(ctx, `INSERT INTO users(id,email,password_hash,role,status) VALUES($1,$2,$3,$4,$5)`, u.ID, strings.ToLower(u.Email), u.PasswordHash, u.Role, u.Status)
	if err != nil && strings.Contains(err.Error(), "duplicate") {
		return entity.ErrConflict
	}
	return err
}
func (r *Repository) CreateRefresh(ctx context.Context, userID uuidv7.UUID, hash string, expires time.Time) (uuidv7.UUID, error) {
	id := uuidv7.New()
	_, err := r.q(ctx).ExecContext(ctx, `INSERT INTO refresh_tokens(id,user_id,token_hash,expires_at) VALUES($1,$2,$3,$4)`, id, userID, hash, expires)
	return id, err
}
func (r *Repository) ConsumeRefresh(ctx context.Context, hash string) (uuidv7.UUID, error) {
	var id, user uuidv7.UUID
	var exp time.Time
	var revoked *time.Time
	err := r.q(ctx).QueryRowContext(ctx, `SELECT id,user_id,expires_at,revoked_at FROM refresh_tokens WHERE token_hash=$1 FOR UPDATE`, hash).Scan(&id, &user, &exp, &revoked)
	if err == sql.ErrNoRows || revoked != nil || time.Now().After(exp) {
		return uuidv7.Nil, entity.ErrUnauthorized
	}
	if err != nil {
		return uuidv7.Nil, err
	}
	_, err = r.q(ctx).ExecContext(ctx, `UPDATE refresh_tokens SET revoked_at=now() WHERE id=$1`, id)
	return user, err
}
func (r *Repository) RevokeRefresh(ctx context.Context, hash string) error {
	_, err := r.q(ctx).ExecContext(ctx, `UPDATE refresh_tokens SET revoked_at=now() WHERE token_hash=$1 AND revoked_at IS NULL`, hash)
	return err
}
func scanTicket(row interface{ Scan(...any) error }) (entity.Ticket, error) {
	var t entity.Ticket
	err := row.Scan(&t.ID, &t.Title, &t.Description, &t.CustomerID, &t.AssigneeID, &t.TeamID, &t.Status, &t.Priority, &t.Category, &t.SLADueAt, &t.SLAPausedAt, &t.SLARemainingSeconds, &t.BreachedAt, &t.CreatedAt, &t.UpdatedAt)
	if err == sql.ErrNoRows {
		return t, entity.ErrNotFound
	}
	return t, err
}

const ticketFields = `id,title,description,customer_id,assignee_id,team_id,status,priority,category,sla_due_at,sla_paused_at,sla_remaining_seconds,breached_at,created_at,updated_at`

func (r *Repository) CreateTicket(ctx context.Context, t entity.Ticket) error {
	_, err := r.q(ctx).ExecContext(ctx, `INSERT INTO tickets(id,title,description,customer_id,status,priority,category,sla_due_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8)`, t.ID, t.Title, t.Description, t.CustomerID, t.Status, t.Priority, t.Category, t.SLADueAt)
	return err
}
func (r *Repository) GetTicket(ctx context.Context, id uuidv7.UUID) (entity.Ticket, error) {
	return scanTicket(r.q(ctx).QueryRowContext(ctx, `SELECT `+ticketFields+` FROM tickets WHERE id=$1`, id))
}
func (r *Repository) ListTickets(ctx context.Context, actor entity.User, limit, offset int, qry string) ([]entity.Ticket, error) {
	sqlq := `SELECT ` + ticketFields + ` FROM tickets WHERE 1=1`
	args := []any{}
	if actor.Role == entity.RoleCustomer {
		args = append(args, actor.ID)
		sqlq += fmt.Sprintf(" AND customer_id=$%d", len(args))
	}
	if qry != "" {
		args = append(args, qry)
		sqlq += fmt.Sprintf(" AND search_vector @@ plainto_tsquery('english',$%d)", len(args))
	}
	args = append(args, limit, offset)
	sqlq += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", len(args)-1, len(args))
	rows, err := r.q(ctx).QueryContext(ctx, sqlq, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []entity.Ticket{}
	for rows.Next() {
		t, e := scanTicket(rows)
		if e != nil {
			return nil, e
		}
		out = append(out, t)
	}
	return out, rows.Err()
}
func (r *Repository) UpdateTicket(ctx context.Context, t entity.Ticket) error {
	_, err := r.q(ctx).ExecContext(ctx, `UPDATE tickets SET status=$2,priority=$3,assignee_id=$4,team_id=$5,sla_due_at=$6,sla_paused_at=$7,sla_remaining_seconds=$8,updated_at=now() WHERE id=$1`, t.ID, t.Status, t.Priority, t.AssigneeID, t.TeamID, t.SLADueAt, t.SLAPausedAt, t.SLARemainingSeconds)
	return err
}
func (r *Repository) AddComment(ctx context.Context, c entity.Comment) error {
	_, err := r.q(ctx).ExecContext(ctx, `INSERT INTO comments(id,ticket_id,author_id,body,visibility) VALUES($1,$2,$3,$4,$5)`, c.ID, c.TicketID, c.AuthorID, c.Body, c.Visibility)
	return err
}
func (r *Repository) Comments(ctx context.Context, ticketID uuidv7.UUID, internal bool) ([]entity.Comment, error) {
	q := `SELECT id,ticket_id,author_id,body,visibility,created_at FROM comments WHERE ticket_id=$1 AND deleted_at IS NULL`
	if !internal {
		q += ` AND visibility='public'`
	}
	rows, err := r.q(ctx).QueryContext(ctx, q+` ORDER BY created_at`, ticketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []entity.Comment{}
	for rows.Next() {
		var c entity.Comment
		if err := rows.Scan(&c.ID, &c.TicketID, &c.AuthorID, &c.Body, &c.Visibility, &c.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
func (r *Repository) Audit(ctx context.Context, ticket, actor uuidv7.UUID, kind string, payload any) error {
	b, _ := json.Marshal(payload)
	_, err := r.q(ctx).ExecContext(ctx, `INSERT INTO audit_events(id,ticket_id,actor_id,event_type,payload) VALUES($1,$2,$3,$4,$5)`, uuidv7.New(), ticket, actor, kind, b)
	return err
}

func (r *Repository) AddStatusHistory(ctx context.Context, ticket, actor uuidv7.UUID, from, to entity.Status, reason string) error {
	_, err := r.q(ctx).ExecContext(ctx, `INSERT INTO ticket_status_history(id,ticket_id,actor_id,from_status,to_status,reason) VALUES($1,$2,$3,$4,$5,$6)`, uuidv7.New(), ticket, actor, from, to, reason)
	return err
}

type TimelineEvent struct {
	ID        uuidv7.UUID     `json:"id"`
	EventType string          `json:"event_type"`
	Payload   json.RawMessage `json:"payload"`
	CreatedAt time.Time       `json:"created_at"`
}

func (r *Repository) Timeline(ctx context.Context, ticket uuidv7.UUID) ([]TimelineEvent, error) {
	rows, err := r.q(ctx).QueryContext(ctx, `SELECT id,event_type,payload,created_at FROM audit_events WHERE ticket_id=$1 ORDER BY created_at`, ticket)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	events := []TimelineEvent{}
	for rows.Next() {
		var e TimelineEvent
		if err := rows.Scan(&e.ID, &e.EventType, &e.Payload, &e.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

func (r *Repository) AddAttachment(ctx context.Context, a entity.Attachment, uploader uuidv7.UUID) error {
	_, err := r.q(ctx).ExecContext(ctx, `INSERT INTO attachments(id,ticket_id,uploaded_by,filename,storage_key,mime_type,size_bytes) VALUES($1,$2,$3,$4,$5,$6,$7)`, a.ID, a.TicketID, uploader, a.Filename, a.StorageKey, a.MIMEType, a.SizeBytes)
	return err
}

func (r *Repository) Attachments(ctx context.Context, ticket uuidv7.UUID) ([]entity.Attachment, error) {
	rows, err := r.q(ctx).QueryContext(ctx, `SELECT id,ticket_id,filename,storage_key,mime_type,size_bytes,created_at FROM attachments WHERE ticket_id=$1 ORDER BY created_at`, ticket)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []entity.Attachment{}
	for rows.Next() {
		var a entity.Attachment
		if err := rows.Scan(&a.ID, &a.TicketID, &a.Filename, &a.StorageKey, &a.MIMEType, &a.SizeBytes, &a.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *Repository) Attachment(ctx context.Context, id uuidv7.UUID) (entity.Attachment, error) {
	var a entity.Attachment
	err := r.q(ctx).QueryRowContext(ctx, `SELECT id,ticket_id,filename,storage_key,mime_type,size_bytes,created_at FROM attachments WHERE id=$1`, id).Scan(&a.ID, &a.TicketID, &a.Filename, &a.StorageKey, &a.MIMEType, &a.SizeBytes, &a.CreatedAt)
	if err == sql.ErrNoRows {
		return a, entity.ErrNotFound
	}
	return a, err
}

func (r *Repository) CreateTeam(ctx context.Context, name string) (entity.Team, error) {
	t := entity.Team{ID: uuidv7.New(), Name: name}
	err := r.q(ctx).QueryRowContext(ctx, `INSERT INTO teams(id,name) VALUES($1,$2) RETURNING created_at`, t.ID, t.Name).Scan(&t.CreatedAt)
	return t, err
}
func (r *Repository) Teams(ctx context.Context) ([]entity.Team, error) {
	rows, err := r.q(ctx).QueryContext(ctx, `SELECT id,name,created_at FROM teams ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []entity.Team{}
	for rows.Next() {
		var t entity.Team
		if err = rows.Scan(&t.ID, &t.Name, &t.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}
func (r *Repository) Agents(ctx context.Context) ([]entity.User, error) {
	rows, err := r.q(ctx).QueryContext(ctx, `SELECT id,email,password_hash,role,status,created_at FROM users WHERE role IN ('agent','admin') ORDER BY email`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []entity.User{}
	for rows.Next() {
		u, e := userScan(rows)
		if e != nil {
			return nil, e
		}
		out = append(out, u)
	}
	return out, rows.Err()
}
func (r *Repository) CreateCannedReply(ctx context.Context, x entity.CannedReply) (entity.CannedReply, error) {
	_, err := r.q(ctx).ExecContext(ctx, `INSERT INTO canned_replies(id,title,body,team_id,created_by) VALUES($1,$2,$3,$4,$5)`, x.ID, x.Title, x.Body, x.TeamID, x.CreatedBy)
	return x, err
}
func (r *Repository) CannedReplies(ctx context.Context) ([]entity.CannedReply, error) {
	rows, err := r.q(ctx).QueryContext(ctx, `SELECT id,title,body,team_id,created_by FROM canned_replies ORDER BY title`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []entity.CannedReply{}
	for rows.Next() {
		var x entity.CannedReply
		if err = rows.Scan(&x.ID, &x.Title, &x.Body, &x.TeamID, &x.CreatedBy); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}
func (r *Repository) CreateTag(ctx context.Context, name string) (entity.Tag, error) {
	x := entity.Tag{ID: uuidv7.New(), Name: name}
	_, err := r.q(ctx).ExecContext(ctx, `INSERT INTO tags(id,name) VALUES($1,$2)`, x.ID, x.Name)
	return x, err
}
func (r *Repository) Tags(ctx context.Context) ([]entity.Tag, error) {
	rows, err := r.q(ctx).QueryContext(ctx, `SELECT id,name FROM tags ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []entity.Tag{}
	for rows.Next() {
		var x entity.Tag
		if err = rows.Scan(&x.ID, &x.Name); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}
func (r *Repository) AttachTag(ctx context.Context, ticket, tag uuidv7.UUID) error {
	_, err := r.q(ctx).ExecContext(ctx, `INSERT INTO ticket_tags(ticket_id,tag_id) VALUES($1,$2) ON CONFLICT DO NOTHING`, ticket, tag)
	return err
}
func (r *Repository) CreateSavedFilter(ctx context.Context, user uuidv7.UUID, name string, query map[string]any) (entity.SavedFilter, error) {
	x := entity.SavedFilter{ID: uuidv7.New(), Name: name, Query: query}
	b, _ := json.Marshal(query)
	_, err := r.q(ctx).ExecContext(ctx, `INSERT INTO saved_filters(id,user_id,name,query) VALUES($1,$2,$3,$4)`, x.ID, user, x.Name, b)
	return x, err
}
func (r *Repository) SavedFilters(ctx context.Context, user uuidv7.UUID) ([]entity.SavedFilter, error) {
	rows, err := r.q(ctx).QueryContext(ctx, `SELECT id,name,query FROM saved_filters WHERE user_id=$1 ORDER BY name`, user)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []entity.SavedFilter{}
	for rows.Next() {
		var x entity.SavedFilter
		var b []byte
		if err = rows.Scan(&x.ID, &x.Name, &b); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(b, &x.Query)
		out = append(out, x)
	}
	return out, rows.Err()
}
func (r *Repository) UpdateCannedReply(ctx context.Context, id uuidv7.UUID, title, body string) error {
	_, err := r.q(ctx).ExecContext(ctx, `UPDATE canned_replies SET title=$2,body=$3,updated_at=now() WHERE id=$1`, id, title, body)
	return err
}
func (r *Repository) DeleteCannedReply(ctx context.Context, id uuidv7.UUID) error {
	_, err := r.q(ctx).ExecContext(ctx, `DELETE FROM canned_replies WHERE id=$1`, id)
	return err
}
func (r *Repository) DeleteTag(ctx context.Context, id uuidv7.UUID) error {
	_, err := r.q(ctx).ExecContext(ctx, `DELETE FROM tags WHERE id=$1`, id)
	return err
}
func (r *Repository) DeleteSavedFilter(ctx context.Context, user, id uuidv7.UUID) error {
	_, err := r.q(ctx).ExecContext(ctx, `DELETE FROM saved_filters WHERE id=$1 AND user_id=$2`, id, user)
	return err
}
func (r *Repository) UpdateTeam(ctx context.Context, id uuidv7.UUID, name string) error {
	_, err := r.q(ctx).ExecContext(ctx, `UPDATE teams SET name=$2 WHERE id=$1`, id, name)
	return err
}
func (r *Repository) DeleteTeam(ctx context.Context, id uuidv7.UUID) error {
	_, err := r.q(ctx).ExecContext(ctx, `DELETE FROM teams WHERE id=$1`, id)
	return err
}
func (r *Repository) UpdateTag(ctx context.Context, id uuidv7.UUID, name string) error {
	_, err := r.q(ctx).ExecContext(ctx, `UPDATE tags SET name=$2 WHERE id=$1`, id, name)
	return err
}
func (r *Repository) UpdateSavedFilter(ctx context.Context, user, id uuidv7.UUID, name string, query map[string]any) error {
	b, _ := json.Marshal(query)
	_, err := r.q(ctx).ExecContext(ctx, `UPDATE saved_filters SET name=$3,query=$4,updated_at=now() WHERE id=$1 AND user_id=$2`, id, user, name, b)
	return err
}
