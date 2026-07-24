package handler

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/XoDeR/customer-support-desk-go/internal/adapter/http/v1/middleware"
	"github.com/XoDeR/customer-support-desk-go/internal/domain/entity"
	"github.com/XoDeR/customer-support-desk-go/pkg/uuidv7"
	"github.com/gin-gonic/gin"
)

var allowedAttachmentTypes = map[string]bool{
	"application/pdf": true, "image/png": true, "image/jpeg": true,
	"image/gif": true, "image/webp": true, "text/plain": true,
}

func (h *Handler) canAccessTicket(c *gin.Context, id uuidv7.UUID) (entity.Ticket, entity.User, error) {
	u, _ := middleware.Current(c)
	t, err := h.Repo.GetTicket(c, id)
	if err != nil {
		return t, u, err
	}
	if u.Role == entity.RoleCustomer && t.CustomerID != u.ID {
		return t, u, entity.ErrNotFound
	}
	return t, u, nil
}

func (h *Handler) UploadAttachment(c *gin.Context) {
	if h.Storage == nil {
		fail(c, errors.New("attachment storage is not configured"))
		return
	}
	id, err := ticketID(c)
	if err != nil {
		fail(c, err)
		return
	}
	_, user, err := h.canAccessTicket(c, id)
	if err != nil {
		fail(c, err)
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		fail(c, errors.New("file is required"))
		return
	}
	if file.Size <= 0 || file.Size > 10*1024*1024 {
		fail(c, errors.New("attachment must be between 1 byte and 10 MiB"))
		return
	}
	contentType := file.Header.Get("Content-Type")
	if !allowedAttachmentTypes[contentType] {
		fail(c, fmt.Errorf("unsupported attachment type %q", contentType))
		return
	}
	existing, err := h.Repo.Attachments(c, id)
	if err != nil {
		fail(c, err)
		return
	}
	if len(existing) >= 10 {
		fail(c, errors.New("ticket attachment limit reached"))
		return
	}
	src, err := file.Open()
	if err != nil {
		fail(c, err)
		return
	}
	defer src.Close()
	attachmentID := uuidv7.New()
	key := filepath.ToSlash(filepath.Join(id.String(), attachmentID.String()+filepath.Ext(file.Filename)))
	if err = h.Storage.Put(c, key, src, file.Size, contentType); err != nil {
		fail(c, err)
		return
	}
	a := entity.Attachment{ID: attachmentID, TicketID: id, Filename: file.Filename, StorageKey: key, MIMEType: contentType, SizeBytes: file.Size}
	if err = h.Repo.AddAttachment(c, a, user.ID); err != nil {
		fail(c, err)
		return
	}
	ok(c, http.StatusCreated, a)
}

func (h *Handler) ListAttachments(c *gin.Context) {
	id, err := ticketID(c)
	if err != nil {
		fail(c, err)
		return
	}
	if _, _, err = h.canAccessTicket(c, id); err != nil {
		fail(c, err)
		return
	}
	items, err := h.Repo.Attachments(c, id)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, http.StatusOK, items)
}

func (h *Handler) DownloadAttachment(c *gin.Context) {
	if h.Storage == nil {
		fail(c, errors.New("attachment storage is not configured"))
		return
	}
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		fail(c, err)
		return
	}
	a, err := h.Repo.Attachment(c, id)
	if err != nil {
		fail(c, err)
		return
	}
	if _, _, err = h.canAccessTicket(c, a.TicketID); err != nil {
		fail(c, err)
		return
	}
	r, err := h.Storage.Get(c, a.StorageKey)
	if err != nil {
		fail(c, err)
		return
	}
	defer r.Close()
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename=%q`, a.Filename))
	c.DataFromReader(http.StatusOK, a.SizeBytes, a.MIMEType, r, nil)
}

func (h *Handler) Timeline(c *gin.Context) {
	id, err := ticketID(c)
	if err != nil {
		fail(c, err)
		return
	}
	if _, _, err = h.canAccessTicket(c, id); err != nil {
		fail(c, err)
		return
	}
	events, err := h.Repo.Timeline(c, id)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, http.StatusOK, events)
}
