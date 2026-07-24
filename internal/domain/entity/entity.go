package entity

import (
	"errors"
	"time"

	"github.com/XoDeR/customer-support-desk-go/pkg/uuidv7"
)

var (
	ErrNotFound           = errors.New("not found")
	ErrConflict           = errors.New("conflict")
	ErrForbidden          = errors.New("forbidden")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidTransition  = errors.New("invalid ticket status transition")
	ErrRateLimited        = errors.New("rate limited")
)

type Role string

const (
	RoleCustomer Role = "customer"
	RoleAgent    Role = "agent"
	RoleAdmin    Role = "admin"
)

type Status string

const (
	StatusOpen     Status = "open"
	StatusPending  Status = "pending"
	StatusResolved Status = "resolved"
	StatusClosed   Status = "closed"
)

type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
	PriorityUrgent Priority = "urgent"
)

type Category string

const (
	CategoryBilling   Category = "billing"
	CategoryTechnical Category = "technical"
	CategoryAccount   Category = "account"
	CategoryOther     Category = "other"
)

type User struct {
	ID           uuidv7.UUID `json:"id"`
	Email        string      `json:"email"`
	PasswordHash string      `json:"-"`
	Role         Role        `json:"role"`
	Status       string      `json:"status"`
	CreatedAt    time.Time   `json:"created_at"`
}
type Ticket struct {
	ID                  uuidv7.UUID  `json:"id"`
	Title               string       `json:"title"`
	Description         string       `json:"description"`
	CustomerID          uuidv7.UUID  `json:"customer_id"`
	AssigneeID          *uuidv7.UUID `json:"assignee_id,omitempty"`
	TeamID              *uuidv7.UUID `json:"team_id,omitempty"`
	Status              Status       `json:"status"`
	Priority            Priority     `json:"priority"`
	Category            Category     `json:"category"`
	SLADueAt            *time.Time   `json:"sla_due_at,omitempty"`
	SLAPausedAt         *time.Time   `json:"sla_paused_at,omitempty"`
	SLARemainingSeconds *int64       `json:"sla_remaining_seconds,omitempty"`
	BreachedAt          *time.Time   `json:"breached_at,omitempty"`
	CreatedAt           time.Time    `json:"created_at"`
	UpdatedAt           time.Time    `json:"updated_at"`
}

func (t *Ticket) TransitionTo(next Status) error {
	if t.Status == next {
		return nil
	}
	allowed := map[Status]map[Status]bool{StatusOpen: {StatusPending: true, StatusResolved: true}, StatusPending: {StatusOpen: true, StatusResolved: true}, StatusResolved: {StatusOpen: true, StatusClosed: true}}
	if !allowed[t.Status][next] {
		return ErrInvalidTransition
	}
	t.Status = next
	return nil
}
func (t *Ticket) PauseSLA(now time.Time) {
	if t.SLADueAt == nil || t.SLAPausedAt != nil {
		return
	}
	v := int64(t.SLADueAt.Sub(now).Seconds())
	if v < 0 {
		v = 0
	}
	t.SLARemainingSeconds = &v
	t.SLAPausedAt = &now
}
func (t *Ticket) ResumeSLA(now time.Time) {
	if t.SLAPausedAt == nil {
		return
	}
	v := int64(0)
	if t.SLARemainingSeconds != nil {
		v = *t.SLARemainingSeconds
	}
	due := now.Add(time.Duration(v) * time.Second)
	t.SLADueAt = &due
	t.SLAPausedAt = nil
	t.SLARemainingSeconds = nil
}

type Comment struct {
	ID         uuidv7.UUID `json:"id"`
	TicketID   uuidv7.UUID `json:"ticket_id"`
	AuthorID   uuidv7.UUID `json:"author_id"`
	Body       string      `json:"body"`
	Visibility string      `json:"visibility"`
	CreatedAt  time.Time   `json:"created_at"`
}
type Attachment struct {
	ID         uuidv7.UUID `json:"id"`
	TicketID   uuidv7.UUID `json:"ticket_id"`
	Filename   string      `json:"filename"`
	MIMEType   string      `json:"mime_type"`
	SizeBytes  int64       `json:"size_bytes"`
	StorageKey string      `json:"-"`
	CreatedAt  time.Time   `json:"created_at"`
}
type Team struct {
	ID        uuidv7.UUID `json:"id"`
	Name      string      `json:"name"`
	CreatedAt time.Time   `json:"created_at"`
}
type Tag struct {
	ID   uuidv7.UUID `json:"id"`
	Name string      `json:"name"`
}
type CannedReply struct {
	ID        uuidv7.UUID  `json:"id"`
	Title     string       `json:"title"`
	Body      string       `json:"body"`
	TeamID    *uuidv7.UUID `json:"team_id,omitempty"`
	CreatedBy uuidv7.UUID  `json:"created_by"`
}
type SavedFilter struct {
	ID    uuidv7.UUID    `json:"id"`
	Name  string         `json:"name"`
	Query map[string]any `json:"query"`
}
type DomainEvent struct {
	Type     string      `json:"type"`
	TicketID uuidv7.UUID `json:"ticket_id"`
	Payload  any         `json:"payload"`
}
