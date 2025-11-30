package middleware

import (
	"net/http"
	"strings"

	"devarminas/kcal-track-api/internal/auth"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func ClerkAuth(v auth.Verifier, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := getSessionToken(c.Request)
		if token == "" {
			logger.Warn("missing session token",
				zap.String("path", c.FullPath()),
				zap.String("method", c.Request.Method),
			)

			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"access": "unauthorized"})
			return
		}

		user, err := v.VerifySession(c.Request.Context(), token)
		if err != nil {
			logger.Warn("failed to verify session",
				zap.Error(err),
				zap.String("path", c.FullPath()),
				zap.String("method", c.Request.Method),
			)

			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"access": "unauthorized"})
			return
		}

		ctx := auth.WithClaims(c.Request.Context(), user)
		c.Request = c.Request.WithContext(ctx)
		logger.Info("authenticated request",
			zap.String("user_id", user.Subject),
			zap.String("path", c.FullPath()),
			zap.String("method", c.Request.Method),
		)
		c.Next()
	}
}

func getSessionToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	cookie, err := r.Cookie("__session")
	if err != nil {
		return ""
	}
	return cookie.Value
}
