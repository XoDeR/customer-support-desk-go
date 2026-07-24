package router

import (
	"github.com/XoDeR/customer-support-desk-go/internal/adapter/http/v1/handler"
	"github.com/XoDeR/customer-support-desk-go/internal/adapter/http/v1/middleware"
	"github.com/XoDeR/customer-support-desk-go/internal/domain/entity"
	"github.com/gin-gonic/gin"
	"net/http"
)

func New(h *handler.Handler, a *middleware.Auth) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.Use(func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "http://localhost:5173" {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})
	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"status": "ok"}}) })
	v := r.Group("/api/v1")
	auth := v.Group("/auth")
	auth.POST("/register", h.Register)
	auth.POST("/login", h.Login)
	auth.POST("/refresh", h.Refresh)
	auth.POST("/logout", h.Logout)
	v.POST("/internal/email-to-ticket", h.EmailToTicket)
	protected := v.Group("")
	protected.Use(a.RequireAuth())
	protected.GET("/me", h.Me)
	protected.GET("/ws", h.WebSocket)
	protected.GET("/attachments/:id/download", h.DownloadAttachment)
	tickets := protected.Group("/tickets")
	tickets.POST("", h.CreateTicket)
	tickets.GET("", h.ListTickets)
	tickets.GET("/search", h.ListTickets)
	tickets.GET("/:id", h.GetTicket)
	tickets.GET("/:id/comments", h.Comments)
	tickets.POST("/:id/comments", h.AddComment)
	tickets.GET("/:id/attachments", h.ListAttachments)
	tickets.POST("/:id/attachments", h.UploadAttachment)
	tickets.GET("/:id/timeline", h.Timeline)
	tickets.POST("/:id/tags", h.AttachTag)
	agent := tickets.Group("")
	agent.Use(a.RequireRoles(entity.RoleAgent, entity.RoleAdmin))
	agent.PATCH("/:id", h.PatchTicket)
	agent.POST("/:id/assign", h.PatchTicket)
	agent.POST("/:id/escalate", h.Escalate)
	ops := protected.Group("")
	ops.Use(a.RequireRoles(entity.RoleAgent, entity.RoleAdmin))
	ops.GET("/agents", h.Agents)
	ops.GET("/teams", h.Teams)
	ops.GET("/canned-replies", h.CannedReplies)
	ops.POST("/canned-replies", h.CreateCannedReply)
	ops.PATCH("/canned-replies/:id", h.UpdateCannedReply)
	ops.DELETE("/canned-replies/:id", h.DeleteCannedReply)
	ops.GET("/tags", h.Tags)
	ops.POST("/tags", h.CreateTag)
	ops.PATCH("/tags/:id", h.UpdateTag)
	ops.DELETE("/tags/:id", h.DeleteTag)
	ops.GET("/saved-filters", h.SavedFilters)
	ops.POST("/saved-filters", h.CreateSavedFilter)
	ops.PATCH("/saved-filters/:id", h.UpdateSavedFilter)
	ops.DELETE("/saved-filters/:id", h.DeleteSavedFilter)
	admin := protected.Group("/teams")
	admin.Use(a.RequireRoles(entity.RoleAdmin))
	admin.POST("", h.CreateTeam)
	admin.PATCH("/:id", h.UpdateTeam)
	admin.DELETE("/:id", h.DeleteTeam)
	return r
}
