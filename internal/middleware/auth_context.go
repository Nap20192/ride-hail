package middleware

import (
	"context"
	"fmt"

	"ride-hail/internal/auth"
	"ride-hail/pkg/uuid"
)

// GetUserIDFromContext extracts the user ID from JWT claims in the request context.
// Returns an error if the context doesn't contain valid claims.
func GetUserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	claims, ok := ctx.Value(UserContextKey).(auth.JWTClaims)
	if !ok {
		return uuid.UUID{}, fmt.Errorf("unauthorized: invalid user context")
	}
	return claims.UserID, nil
}

// GetUserRoleFromContext extracts the user role from JWT claims in the request context.
// Returns an error if the context doesn't contain valid claims.
func GetUserRoleFromContext(ctx context.Context) (string, error) {
	claims, ok := ctx.Value(UserContextKey).(auth.JWTClaims)
	if !ok {
		return "", fmt.Errorf("unauthorized: invalid user context")
	}
	return claims.Role, nil
}

// GetClaimsFromContext extracts the full JWT claims from the request context.
// Returns an error if the context doesn't contain valid claims.
func GetClaimsFromContext(ctx context.Context) (auth.JWTClaims, error) {
	claims, ok := ctx.Value(UserContextKey).(auth.JWTClaims)
	if !ok {
		return auth.JWTClaims{}, fmt.Errorf("unauthorized: invalid user context")
	}
	return claims, nil
}
