package handler

import (
	"errors"
	"net/http"

	"github.com/XoDeR/customer-support-desk-go/internal/adapter/http/v1/middleware"
	"github.com/XoDeR/customer-support-desk-go/internal/domain/entity"
	"github.com/XoDeR/customer-support-desk-go/pkg/uuidv7"
	"github.com/gin-gonic/gin"
)

func (h *Handler) Teams(c *gin.Context) {
	x, e := h.Repo.Teams(c)
	if e != nil {
		fail(c, e)
		return
	}
	ok(c, http.StatusOK, x)
}
func (h *Handler) CreateTeam(c *gin.Context) {
	var r struct {
		Name string `json:"name" binding:"required"`
	}
	if c.ShouldBindJSON(&r) != nil {
		fail(c, errors.New("invalid request"))
		return
	}
	x, e := h.Repo.CreateTeam(c, r.Name)
	if e != nil {
		fail(c, e)
		return
	}
	ok(c, http.StatusCreated, x)
}
func (h *Handler) UpdateTeam(c *gin.Context) { id,e:=uuidv7.Parse(c.Param("id"));var r struct{Name string `json:"name" binding:"required"`};if e==nil&&c.ShouldBindJSON(&r)==nil{e=h.Repo.UpdateTeam(c,id,r.Name)};if e!=nil{fail(c,e);return};ok(c,http.StatusOK,gin.H{}) }
func (h *Handler) DeleteTeam(c *gin.Context) { id,e:=uuidv7.Parse(c.Param("id"));if e==nil{e=h.Repo.DeleteTeam(c,id)};if e!=nil{fail(c,e);return};ok(c,http.StatusOK,gin.H{}) }
func (h *Handler) Agents(c *gin.Context) {
	x, e := h.Repo.Agents(c)
	if e != nil {
		fail(c, e)
		return
	}
	ok(c, http.StatusOK, x)
}

func (h *Handler) CannedReplies(c *gin.Context) {
	x, e := h.Repo.CannedReplies(c)
	if e != nil {
		fail(c, e)
		return
	}
	ok(c, http.StatusOK, x)
}
func (h *Handler) CreateCannedReply(c *gin.Context) {
	u, _ := middleware.Current(c)
	var r struct {
		Title  string       `json:"title" binding:"required"`
		Body   string       `json:"body" binding:"required"`
		TeamID *uuidv7.UUID `json:"team_id"`
	}
	if c.ShouldBindJSON(&r) != nil {
		fail(c, errors.New("invalid request"))
		return
	}
	x, e := h.Repo.CreateCannedReply(c, entity.CannedReply{ID: uuidv7.New(), Title: r.Title, Body: r.Body, TeamID: r.TeamID, CreatedBy: u.ID})
	if e != nil {
		fail(c, e)
		return
	}
	ok(c, http.StatusCreated, x)
}
func (h *Handler) UpdateCannedReply(c *gin.Context) {
	id, e := uuidv7.Parse(c.Param("id"))
	if e != nil {
		fail(c, e)
		return
	}
	var r struct {
		Title string `json:"title"`
		Body  string `json:"body"`
	}
	if c.ShouldBindJSON(&r) != nil {
		fail(c, errors.New("invalid request"))
		return
	}
	if e = h.Repo.UpdateCannedReply(c, id, r.Title, r.Body); e != nil {
		fail(c, e)
		return
	}
	ok(c, http.StatusOK, gin.H{})
}
func (h *Handler) DeleteCannedReply(c *gin.Context) {
	id, e := uuidv7.Parse(c.Param("id"))
	if e == nil {
		e = h.Repo.DeleteCannedReply(c, id)
	}
	if e != nil {
		fail(c, e)
		return
	}
	ok(c, http.StatusOK, gin.H{})
}

func (h *Handler) Tags(c *gin.Context) {
	x, e := h.Repo.Tags(c)
	if e != nil {
		fail(c, e)
		return
	}
	ok(c, http.StatusOK, x)
}
func (h *Handler) CreateTag(c *gin.Context) {
	var r struct {
		Name string `json:"name" binding:"required"`
	}
	if c.ShouldBindJSON(&r) != nil {
		fail(c, errors.New("invalid request"))
		return
	}
	x, e := h.Repo.CreateTag(c, r.Name)
	if e != nil {
		fail(c, e)
		return
	}
	ok(c, http.StatusCreated, x)
}
func (h *Handler) DeleteTag(c *gin.Context) {
	id, e := uuidv7.Parse(c.Param("id"))
	if e == nil {
		e = h.Repo.DeleteTag(c, id)
	}
	if e != nil {
		fail(c, e)
		return
	}
	ok(c, http.StatusOK, gin.H{})
}
func (h *Handler) UpdateTag(c *gin.Context) { id,e:=uuidv7.Parse(c.Param("id"));var r struct{Name string `json:"name" binding:"required"`};if e==nil&&c.ShouldBindJSON(&r)==nil{e=h.Repo.UpdateTag(c,id,r.Name)};if e!=nil{fail(c,e);return};ok(c,http.StatusOK,gin.H{}) }
func (h *Handler) AttachTag(c *gin.Context) {
	ticket, e := ticketID(c)
	if e != nil {
		fail(c, e)
		return
	}
	var r struct {
		TagID uuidv7.UUID `json:"tag_id" binding:"required"`
	}
	if c.ShouldBindJSON(&r) != nil {
		fail(c, errors.New("invalid request"))
		return
	}
	if e = h.Repo.AttachTag(c, ticket, r.TagID); e != nil {
		fail(c, e)
		return
	}
	ok(c, http.StatusCreated, gin.H{})
}

func (h *Handler) SavedFilters(c *gin.Context) {
	u, _ := middleware.Current(c)
	x, e := h.Repo.SavedFilters(c, u.ID)
	if e != nil {
		fail(c, e)
		return
	}
	ok(c, http.StatusOK, x)
}
func (h *Handler) CreateSavedFilter(c *gin.Context) {
	u, _ := middleware.Current(c)
	var r struct {
		Name  string         `json:"name" binding:"required"`
		Query map[string]any `json:"query"`
	}
	if c.ShouldBindJSON(&r) != nil {
		fail(c, errors.New("invalid request"))
		return
	}
	x, e := h.Repo.CreateSavedFilter(c, u.ID, r.Name, r.Query)
	if e != nil {
		fail(c, e)
		return
	}
	ok(c, http.StatusCreated, x)
}
func (h *Handler) DeleteSavedFilter(c *gin.Context) {
	u, _ := middleware.Current(c)
	id, e := uuidv7.Parse(c.Param("id"))
	if e == nil {
		e = h.Repo.DeleteSavedFilter(c, u.ID, id)
	}
	if e != nil {
		fail(c, e)
		return
	}
	ok(c, http.StatusOK, gin.H{})
}
func (h *Handler) UpdateSavedFilter(c *gin.Context) { u,_:=middleware.Current(c);id,e:=uuidv7.Parse(c.Param("id"));var r struct{Name string `json:"name" binding:"required"`;Query map[string]any `json:"query"`};if e==nil&&c.ShouldBindJSON(&r)==nil{e=h.Repo.UpdateSavedFilter(c,u.ID,id,r.Name,r.Query)};if e!=nil{fail(c,e);return};ok(c,http.StatusOK,gin.H{}) }
