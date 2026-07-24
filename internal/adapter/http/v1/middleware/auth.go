package middleware

import (
	"github.com/XoDeR/customer-support-desk-go/internal/domain/entity"
	jwtpkg "github.com/XoDeR/customer-support-desk-go/pkg/jwt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

const userKey = "current_user"

type Auth struct{ jwt *jwtpkg.Manager }

func NewAuth(j *jwtpkg.Manager) *Auth { return &Auth{j} }
func (a *Auth) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		v := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
		if v == "" || v == c.GetHeader("Authorization") {
			// Browsers cannot set Authorization during a WebSocket handshake.
			// Limit token query-string support to the WebSocket endpoint.
			if c.Request.URL.Path == "/api/v1/ws" {
				v = c.Query("token")
			}
		}
		if v == "" || v == c.GetHeader("Authorization") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"success": false, "error": "authentication required"})
			return
		}
		claims, e := a.jwt.ValidateToken(v)
		if e != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"success": false, "error": "invalid token"})
			return
		}
		c.Set(userKey, entity.User{ID: claims.UserID, Email: claims.Email, Role: entity.Role(claims.Role), Status: "active"})
		c.Next()
	}
}
func (a *Auth) RequireRoles(roles ...entity.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		u, ok := Current(c)
		if !ok {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		for _, r := range roles {
			if u.Role == r {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"success": false, "error": "forbidden"})
	}
}
func Current(c *gin.Context) (entity.User, bool) {
	v, ok := c.Get(userKey)
	if !ok {
		return entity.User{}, false
	}
	u, ok := v.(entity.User)
	return u, ok
}
