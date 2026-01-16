package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"ride-hail/internal/auth"
	"ride-hail/internal/shared/core"
)

type contextKey string

const (
	UserContextKey contextKey = "claims"
)

func AuthMiddleware(authService auth.AuthService, role core.UserRole) Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			authHeader := r.Header.Get("Authorization")

			if authHeader == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			if !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			token := strings.TrimPrefix(authHeader, "Bearer ")

			if token == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			session, err := authService.ParseToken(token)
			if err != nil {
				if err == auth.ErrUnauthorized {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}

				slog.Error("Auth middleware error parsing token", slog.String("error", err.Error()))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			if session.Role != role.String() {
				http.Error(w, "user "+session.Role+" is not authorized", http.StatusForbidden)
				return
			}

			ctx = context.WithValue(ctx, UserContextKey, session)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserFromContext(ctx context.Context) (auth.JWTClaims, bool) {
	user, ok := ctx.Value(UserContextKey).(auth.JWTClaims)
	return user, ok
}
