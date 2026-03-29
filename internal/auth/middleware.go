package auth

import (
	"context"
	"errors"
	"fileServer/internal/logger"
	"net/http"
	"strings"
)

type contextKey string

const ClaimsKey contextKey = "claims"

func JWTMiddleware(manager *JWTManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr, err := extractBearerToken(r)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			claims, err := manager.Verify(tokenStr)
			if err != nil {
				logger.Debugf("JWTMiddleware: invalid token on %s %s: %v", r.Method, r.URL.Path, err)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			logger.Debugf("JWTMiddleware: user %s (role: %s) -> %s %s", claims.Username, claims.Role, r.Method, r.URL.Path)
			ctx := context.WithValue(r.Context(), ClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func extractBearerToken(r *http.Request) (string, error) {
	header := r.Header.Get("Authorization")
	if header == "" {
		return "", errors.New("missing authorization header")
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", errors.New("invalid authorization header format")
	}

	return parts[1], nil
}
