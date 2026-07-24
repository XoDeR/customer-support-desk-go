package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/XoDeR/customer-support-desk-go/internal/adapter/repository/postgres"
	"github.com/XoDeR/customer-support-desk-go/internal/domain/entity"
	"github.com/XoDeR/customer-support-desk-go/internal/infrastructure/database"
	jwtpkg "github.com/XoDeR/customer-support-desk-go/pkg/jwt"
	"github.com/XoDeR/customer-support-desk-go/pkg/uuidv7"
	"golang.org/x/crypto/argon2"
)

type Publisher interface {
	Publish(context.Context, entity.DomainEvent) error
}
type EscalationQueue interface {
	EnqueueEscalation(context.Context, entity.DomainEvent) error
}
type noopPublisher struct{}

func (noopPublisher) Publish(context.Context, entity.DomainEvent) error { return nil }

type Auth struct {
	repo *postgres.Repository
	tx   database.TransactionManager
	jwt  *jwtpkg.Manager
}

func NewAuth(r *postgres.Repository, tx database.TransactionManager, j *jwtpkg.Manager) *Auth {
	return &Auth{r, tx, j}
}
func HashPassword(p string) (string, error) {
	salt := make([]byte, 16)
	if _, e := rand.Read(salt); e != nil {
		return "", e
	}
	h := argon2.IDKey([]byte(p), salt, 1, 65536, 4, 32)
	return fmt.Sprintf("argon2id$v=19$m=65536,t=1,p=4$%s$%s", base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(h)), nil
}
func VerifyPassword(encoded, p string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 5 {
		return false
	}
	salt, e := base64.RawStdEncoding.DecodeString(parts[3])
	if e != nil {
		return false
	}
	want, e := base64.RawStdEncoding.DecodeString(parts[4])
	if e != nil {
		return false
	}
	got := argon2.IDKey([]byte(p), salt, 1, 65536, 4, uint32(len(want)))
	return subtle.ConstantTimeCompare(got, want) == 1
}
func tokenHash(s string) string {
	h := sha256.Sum256([]byte(s))
	return base64.RawStdEncoding.EncodeToString(h[:])
}
func randomToken() (string, error) {
	b := make([]byte, 48)
	_, e := rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b), e
}

type Tokens struct {
	AccessToken     string    `json:"access_token"`
	RefreshToken    string    `json:"refresh_token"`
	AccessExpiresAt time.Time `json:"access_expires_at"`
}

func (a *Auth) Register(ctx context.Context, email, password string) (entity.User, error) {
	if len(password) < 8 {
		return entity.User{}, fmt.Errorf("password must have at least 8 characters")
	}
	h, e := HashPassword(password)
	if e != nil {
		return entity.User{}, e
	}
	u := entity.User{ID: uuidv7.New(), Email: email, PasswordHash: h, Role: entity.RoleCustomer, Status: "active"}
	return u, a.repo.CreateUser(ctx, u)
}
func (a *Auth) issue(ctx context.Context, u entity.User) (Tokens, error) {
	access, exp, e := a.jwt.GenerateAccessToken(u.ID, u.Email, string(u.Role))
	if e != nil {
		return Tokens{}, e
	}
	refresh, e := randomToken()
	if e != nil {
		return Tokens{}, e
	}
	_, e = a.repo.CreateRefresh(ctx, u.ID, tokenHash(refresh), time.Now().Add(a.jwt.RefreshTTL()))
	return Tokens{access, refresh, exp}, e
}
func (a *Auth) Login(ctx context.Context, email, password string) (Tokens, entity.User, error) {
	u, e := a.repo.GetUserByEmail(ctx, email)
	if e != nil || u.Status != "active" || !VerifyPassword(u.PasswordHash, password) {
		return Tokens{}, entity.User{}, entity.ErrInvalidCredentials
	}
	t, e := a.issue(ctx, u)
	return t, u, e
}
func (a *Auth) Refresh(ctx context.Context, refresh string) (Tokens, error) {
	var u entity.User
	var out Tokens
	e := a.tx.WithTransaction(ctx, func(tx context.Context) error {
		uid, e := a.repo.ConsumeRefresh(tx, tokenHash(refresh))
		if e != nil {
			return e
		}
		u, e = a.repo.GetUser(tx, uid)
		if e != nil {
			return e
		}
		out, e = a.issue(tx, u)
		return e
	})
	return out, e
}
func (a *Auth) Logout(ctx context.Context, refresh string) error {
	return a.repo.RevokeRefresh(ctx, tokenHash(refresh))
}

type Tickets struct {
	repo  *postgres.Repository
	tx    database.TransactionManager
	pub   Publisher
	queue EscalationQueue
}

func NewTickets(r *postgres.Repository, tx database.TransactionManager, p Publisher) *Tickets {
	if p == nil {
		p = noopPublisher{}
	}
	t := &Tickets{repo: r, tx: tx, pub: p}
	if queue, ok := p.(EscalationQueue); ok {
		t.queue = queue
	}
	return t
}
func (u *Tickets) authorized(ctx context.Context, actor entity.User, id uuidv7.UUID) (entity.Ticket, error) {
	t, e := u.repo.GetTicket(ctx, id)
	if e != nil {
		return t, e
	}
	if actor.Role == entity.RoleCustomer && t.CustomerID != actor.ID {
		return t, entity.ErrNotFound
	}
	return t, nil
}
func duration(p entity.Priority) time.Duration {
	switch p {
	case entity.PriorityLow:
		return 72 * time.Hour
	case entity.PriorityHigh:
		return 8 * time.Hour
	case entity.PriorityUrgent:
		return 2 * time.Hour
	default:
		return 24 * time.Hour
	}
}
func (u *Tickets) Create(ctx context.Context, actor entity.User, title, description string, category entity.Category, priority entity.Priority) (entity.Ticket, error) {
	if priority == "" {
		priority = entity.PriorityMedium
	}
	if category == "" {
		category = entity.CategoryOther
	}
	now := time.Now()
	due := now.Add(duration(priority))
	t := entity.Ticket{ID: uuidv7.New(), Title: title, Description: description, CustomerID: actor.ID, Status: entity.StatusOpen, Priority: priority, Category: category, SLADueAt: &due}
	e := u.tx.WithTransaction(ctx, func(c context.Context) error {
		if e := u.repo.CreateTicket(c, t); e != nil {
			return e
		}
		return u.repo.Audit(c, t.ID, actor.ID, "ticket.created", t)
	})
	if e == nil {
		e = u.pub.Publish(ctx, entity.DomainEvent{Type: "ticket.created", TicketID: t.ID, Payload: t})
	}
	return t, e
}
func (u *Tickets) Update(ctx context.Context, actor entity.User, id uuidv7.UUID, status entity.Status, priority entity.Priority, assignee, team *uuidv7.UUID) (entity.Ticket, error) {
	if actor.Role == entity.RoleCustomer {
		return entity.Ticket{}, entity.ErrForbidden
	}
	var t entity.Ticket
	e := u.tx.WithTransaction(ctx, func(c context.Context) error {
		var e error
		t, e = u.repo.GetTicket(c, id)
		if e != nil {
			return e
		}
		old := t.Status
		if status != "" {
			if e = t.TransitionTo(status); e != nil {
				return e
			}
			if status == entity.StatusPending {
				t.PauseSLA(time.Now())
			}
			if old == entity.StatusPending && status != entity.StatusPending {
				t.ResumeSLA(time.Now())
			}
			if old != status {
				if e = u.repo.AddStatusHistory(c, id, actor.ID, old, status, "status updated"); e != nil {
					return e
				}
			}
		}
		if priority != "" && priority != t.Priority {
			t.Priority = priority
			d := time.Now().Add(duration(priority))
			t.SLADueAt = &d
			t.SLAPausedAt = nil
			t.SLARemainingSeconds = nil
		}
		if assignee != nil {
			t.AssigneeID = assignee
		}
		if team != nil {
			t.TeamID = team
		}
		if e = u.repo.UpdateTicket(c, t); e != nil {
			return e
		}
		return u.repo.Audit(c, id, actor.ID, "ticket.updated", t)
	})
	if e == nil {
		e = u.pub.Publish(ctx, entity.DomainEvent{Type: "ticket.updated", TicketID: id, Payload: t})
	}
	return t, e
}
func (u *Tickets) Comment(ctx context.Context, actor entity.User, id uuidv7.UUID, body, visibility string) (entity.Comment, error) {
	if _, e := u.authorized(ctx, actor, id); e != nil {
		return entity.Comment{}, e
	}
	if visibility == "internal" && actor.Role == entity.RoleCustomer {
		return entity.Comment{}, entity.ErrForbidden
	}
	if visibility == "" {
		visibility = "public"
	}
	c := entity.Comment{ID: uuidv7.New(), TicketID: id, AuthorID: actor.ID, Body: body, Visibility: visibility}
	e := u.repo.AddComment(ctx, c)
	if e == nil {
		e = u.pub.Publish(ctx, entity.DomainEvent{Type: "comment.created", TicketID: id, Payload: map[string]any{"comment_id": c.ID, "visibility": visibility}})
	}
	return c, e
}
func (u *Tickets) Escalate(ctx context.Context, actor entity.User, id uuidv7.UUID) (entity.Ticket, error) {
	t, e := u.authorized(ctx, actor, id)
	if e != nil {
		return t, e
	}
	next := map[entity.Priority]entity.Priority{entity.PriorityLow: entity.PriorityMedium, entity.PriorityMedium: entity.PriorityHigh, entity.PriorityHigh: entity.PriorityUrgent, entity.PriorityUrgent: entity.PriorityUrgent}[t.Priority]
	updated, err := u.Update(ctx, actor, id, "", next, nil, nil)
	if err != nil {
		return updated, err
	}
	if u.queue != nil {
		err = u.queue.EnqueueEscalation(ctx, entity.DomainEvent{Type: "ticket.escalated", TicketID: id, Payload: map[string]any{"actor_id": actor.ID, "priority": next}})
	}
	return updated, err
}
