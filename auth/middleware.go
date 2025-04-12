package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/rb4807/Golang-Utlis-Postgresql/utils"
)

// AuthMiddleware is a middleware function to protect routes
func (s *Service) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			utils.SendJSONError(w, "Authorization header is required", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.SendJSONError(w, "Authorization header format must be Bearer <token>", http.StatusUnauthorized)
			return
		}

		claims, err := s.VerifyJWT(parts[1])
		if err != nil {
			utils.SendJSONError(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// AdminMiddleware is a middleware function to protect admin routes
func (s *Service) AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// First apply auth middleware
		s.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if user is staff or superuser
			claims := r.Context().Value(UserContextKey).(*TokenClaims)
			if !claims.IsSuperuser {
				http.Error(w, "Admin access required", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})).ServeHTTP(w, r)
	})
}

// SuperuserMiddleware is a middleware function to protect superuser routes
func (s *Service) SuperuserMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// First apply auth middleware
		s.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if user is superuser
			claims := r.Context().Value(UserContextKey).(*TokenClaims)
			if !claims.IsSuperuser {
				http.Error(w, "Superuser access required", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})).ServeHTTP(w, r)
	})
}

// RequireAuth is a middleware generator that can be used to protect routes with custom logic
func (s *Service) RequireAuth(checkFunc func(*TokenClaims) bool, message string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			s.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				claims := r.Context().Value(UserContextKey).(*TokenClaims)
				if !checkFunc(claims) {
					http.Error(w, message, http.StatusForbidden)
					return
				}
				next.ServeHTTP(w, r)
			})).ServeHTTP(w, r)
		})
	}
}