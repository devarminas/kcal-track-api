package middleware

import (
	"devarminas/kcal-track-api/internal/auth"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

func ClerkAuth(v auth.Verifier, logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := getSessionToken(r)
			if token == "" {
				logger.Warn("missing session token",
					zap.String("path", r.URL.Path),
					zap.String("method", r.Method),
				)
				unauthorized(w)
				return
			}

			user, err := v.VerifySession(r.Context(), token)
			if err != nil {
				logger.Warn("failed to verify session",
					zap.Error(err),
					zap.String("path", r.URL.Path),
					zap.String("method", r.Method),
				)
				unauthorized(w)
				return
			}

			ctx := auth.WithClaims(r.Context(), user)
			logger.Info("authenticated request",
				zap.String("user_id", user.Subject),
				zap.String("path", r.URL.Path),
				zap.String("method", r.Method),
			)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func unauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"access":"unauthorized"}`))
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
