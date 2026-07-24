package repository

import (
	"context"

	"github.com/XoDeR/customer-support-desk-go/internal/domain/entity"
	"github.com/XoDeR/customer-support-desk-go/pkg/uuidv7"
)

// These ports keep use-case contracts independent of PostgreSQL/sqlc details.
type UserRepository interface {
	GetUser(context.Context, uuidv7.UUID) (entity.User, error)
	GetUserByEmail(context.Context, string) (entity.User, error)
	CreateUser(context.Context, entity.User) error
}
type TicketRepository interface {
	CreateTicket(context.Context, entity.Ticket) error
	GetTicket(context.Context, uuidv7.UUID) (entity.Ticket, error)
	UpdateTicket(context.Context, entity.Ticket) error
}
type CommentRepository interface {
	AddComment(context.Context, entity.Comment) error
	Comments(context.Context, uuidv7.UUID, bool) ([]entity.Comment, error)
}
