package handler

import (
	"errors"
	"github.com/XoDeR/customer-support-desk-go/internal/adapter/http/v1/middleware"
	"github.com/XoDeR/customer-support-desk-go/internal/adapter/repository/postgres"
	"github.com/XoDeR/customer-support-desk-go/internal/adapter/storage"
	"github.com/XoDeR/customer-support-desk-go/internal/adapter/ws"
	"github.com/XoDeR/customer-support-desk-go/internal/domain/entity"
	redisinfrastructure "github.com/XoDeR/customer-support-desk-go/internal/infrastructure/redis"
	"github.com/XoDeR/customer-support-desk-go/internal/usecase"
	"github.com/XoDeR/customer-support-desk-go/pkg/uuidv7"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Handler struct {
	Auth          *usecase.Auth
	Tickets       *usecase.Tickets
	Repo          *postgres.Repository
	InternalToken string
	Storage       storage.ObjectStorage
	Limiter       *redisinfrastructure.Limiter
	Hub           *ws.Hub
}

func (h *Handler) WebSocket(c *gin.Context) {
	if h.Hub == nil {
		fail(c, errors.New("realtime is not configured"))
		return
	}
	u, _ := middleware.Current(c)
	h.Hub.ServeHTTP(c.Writer, c.Request, u.Role)
}

func ok(c *gin.Context, status int, data any) { c.JSON(status, gin.H{"success": true, "data": data}) }
func fail(c *gin.Context, e error) {
	status := http.StatusBadRequest
	switch {
	case errors.Is(e, entity.ErrUnauthorized), errors.Is(e, entity.ErrInvalidCredentials):
		status = 401
	case errors.Is(e, entity.ErrForbidden):
		status = 403
	case errors.Is(e, entity.ErrNotFound):
		status = 404
	case errors.Is(e, entity.ErrConflict):
		status = 409
	case errors.Is(e, entity.ErrRateLimited):
		status = 429
	}
	c.JSON(status, gin.H{"success": false, "error": e.Error()})
}
func (h *Handler) Register(c *gin.Context) {
	var r struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
	if c.ShouldBindJSON(&r) != nil {
		fail(c, errors.New("invalid request"))
		return
	}
	u, e := h.Auth.Register(c, r.Email, r.Password)
	if e != nil {
		fail(c, e)
		return
	}
	ok(c, 201, u)
}
func (h *Handler) Login(c *gin.Context) {
	var r struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if c.ShouldBindJSON(&r) != nil {
		fail(c, errors.New("invalid request"))
		return
	}
	if h.Limiter != nil {
		allowed, err := h.Limiter.Allow(c, "rate:login:"+c.ClientIP()+":"+strings.ToLower(r.Email), 10, 15*time.Minute)
		if err != nil {
			fail(c, err)
			return
		}
		if !allowed {
			fail(c, entity.ErrRateLimited)
			return
		}
	}
	t, _, e := h.Auth.Login(c, r.Email, r.Password)
	if e != nil {
		fail(c, e)
		return
	}
	ok(c, 200, t)
}
func (h *Handler) Refresh(c *gin.Context) {
	var r struct {
		RefreshToken string `json:"refresh_token"`
	}
	if c.ShouldBindJSON(&r) != nil {
		fail(c, errors.New("invalid request"))
		return
	}
	t, e := h.Auth.Refresh(c, r.RefreshToken)
	if e != nil {
		fail(c, e)
		return
	}
	ok(c, 200, t)
}
func (h *Handler) Logout(c *gin.Context) {
	var r struct {
		RefreshToken string `json:"refresh_token"`
	}
	_ = c.ShouldBindJSON(&r)
	if e := h.Auth.Logout(c, r.RefreshToken); e != nil {
		fail(c, e)
		return
	}
	ok(c, 200, gin.H{})
}
func (h *Handler) Me(c *gin.Context) {
	u, _ := middleware.Current(c)
	dbu, e := h.Repo.GetUser(c, u.ID)
	if e != nil {
		fail(c, e)
		return
	}
	ok(c, 200, dbu)
}
func ticketID(c *gin.Context) (uuidv7.UUID, error) { return uuidv7.Parse(c.Param("id")) }
func (h *Handler) CreateTicket(c *gin.Context) {
	u, _ := middleware.Current(c)
	var r struct {
		Title       string          `json:"title" binding:"required"`
		Description string          `json:"description" binding:"required"`
		Category    entity.Category `json:"category"`
		Priority    entity.Priority `json:"priority"`
	}
	if c.ShouldBindJSON(&r) != nil {
		fail(c, errors.New("invalid request"))
		return
	}
	if h.Limiter != nil {
		allowed, err := h.Limiter.Allow(c, "rate:ticket:"+u.ID.String(), 10, 10*time.Minute)
		if err != nil {
			fail(c, err)
			return
		}
		if !allowed {
			fail(c, entity.ErrRateLimited)
			return
		}
	}
	t, e := h.Tickets.Create(c, u, r.Title, r.Description, r.Category, r.Priority)
	if e != nil {
		fail(c, e)
		return
	}
	ok(c, 201, t)
}
func (h *Handler) ListTickets(c *gin.Context) {
	u, _ := middleware.Current(c)
	l, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	o, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if l < 1 || l > 100 {
		l = 50
	}
	ts, e := h.Repo.ListTickets(c, u, l, o, c.Query("q"))
	if e != nil {
		fail(c, e)
		return
	}
	ok(c, 200, ts)
}
func (h *Handler) GetTicket(c *gin.Context) {
	u, _ := middleware.Current(c)
	id, e := ticketID(c)
	if e == nil {
		var t entity.Ticket
		t, e = h.Repo.GetTicket(c, id)
		if e == nil && u.Role == entity.RoleCustomer && t.CustomerID != u.ID {
			e = entity.ErrNotFound
		}
		if e == nil {
			ok(c, 200, t)
			return
		}
	}
	fail(c, e)
}
func (h *Handler) PatchTicket(c *gin.Context) {
	u, _ := middleware.Current(c)
	id, e := ticketID(c)
	if e != nil {
		fail(c, e)
		return
	}
	var r struct {
		Status     entity.Status   `json:"status"`
		Priority   entity.Priority `json:"priority"`
		AssigneeID *uuidv7.UUID    `json:"assignee_id"`
		TeamID     *uuidv7.UUID    `json:"team_id"`
	}
	if c.ShouldBindJSON(&r) != nil {
		fail(c, errors.New("invalid request"))
		return
	}
	t, e := h.Tickets.Update(c, u, id, r.Status, r.Priority, r.AssigneeID, r.TeamID)
	if e != nil {
		fail(c, e)
		return
	}
	ok(c, 200, t)
}
func (h *Handler) Comments(c *gin.Context) {
	u, _ := middleware.Current(c)
	id, e := ticketID(c)
	if e != nil {
		fail(c, e)
		return
	}
	t, e := h.Repo.GetTicket(c, id)
	if e != nil || (u.Role == entity.RoleCustomer && t.CustomerID != u.ID) {
		fail(c, entity.ErrNotFound)
		return
	}
	cs, e := h.Repo.Comments(c, id, u.Role != entity.RoleCustomer)
	if e != nil {
		fail(c, e)
		return
	}
	ok(c, 200, cs)
}
func (h *Handler) AddComment(c *gin.Context) {
	u, _ := middleware.Current(c)
	id, e := ticketID(c)
	if e != nil {
		fail(c, e)
		return
	}
	var r struct {
		Body       string `json:"body" binding:"required"`
		Visibility string `json:"visibility"`
	}
	if c.ShouldBindJSON(&r) != nil {
		fail(c, errors.New("invalid request"))
		return
	}
	if h.Limiter != nil {
		allowed, err := h.Limiter.Allow(c, "rate:reply:"+u.ID.String(), 60, 10*time.Minute)
		if err != nil {
			fail(c, err)
			return
		}
		if !allowed {
			fail(c, entity.ErrRateLimited)
			return
		}
	}
	x, e := h.Tickets.Comment(c, u, id, r.Body, r.Visibility)
	if e != nil {
		fail(c, e)
		return
	}
	ok(c, 201, x)
}
func (h *Handler) Escalate(c *gin.Context) {
	u, _ := middleware.Current(c)
	id, e := ticketID(c)
	if e == nil {
		t, e := h.Tickets.Escalate(c, u, id)
		if e == nil {
			ok(c, 200, t)
			return
		}
	}
	fail(c, e)
}
func (h *Handler) EmailToTicket(c *gin.Context) {
	if h.InternalToken == "" || c.GetHeader("X-Internal-Token") != h.InternalToken {
		fail(c, entity.ErrForbidden)
		return
	}
	var r struct {
		FromEmail string          `json:"from_email"`
		Subject   string          `json:"subject"`
		Body      string          `json:"body"`
		Category  entity.Category `json:"category"`
		Priority  entity.Priority `json:"priority"`
	}
	if c.ShouldBindJSON(&r) != nil {
		fail(c, errors.New("invalid request"))
		return
	}
	u, e := h.Repo.GetUserByEmail(c, r.FromEmail)
	if errors.Is(e, entity.ErrNotFound) {
		u, e = h.Auth.Register(c, r.FromEmail, "unusable-"+uuidv7.New().String())
	}
	if e != nil {
		fail(c, e)
		return
	}
	t, e := h.Tickets.Create(c, u, r.Subject, r.Body, r.Category, r.Priority)
	if e != nil {
		fail(c, e)
		return
	}
	ok(c, 201, t)
}
