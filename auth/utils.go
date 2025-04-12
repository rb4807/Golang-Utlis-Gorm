package auth

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

// Context key for storing user info in request context
type contextKey string
const UserContextKey contextKey = "user"

// NewService creates a new authentication service
func NewService(config Config) (*Service, error) {
	if config.JWTSecret == "" {
		return nil, errors.New("JWT secret is required")
	}
	if config.DB == nil {
		return nil, errors.New("DB connection is required")
	}
	
	validate := validator.New()
	
	return &Service{
		config:    config,
		validator: validate,
	}, nil
}

// validateData validates a struct using the validator
func (s *Service) validateData(data interface{}) error {
	if v, ok := s.validator.(*validator.Validate); ok {
		return v.Struct(data)
	}
	return errors.New("validator not initialized")
}

// HashPassword hashes a password using bcrypt
func (s *Service) HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// VerifyPassword checks if a password matches the hash
func (s *Service) VerifyPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// generateRandomOTP creates a random numeric OTP of specified length
func (s *Service) generateRandomOTP(length int) (string, error) {
	otp := ""
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		otp += fmt.Sprintf("%d", n.Int64())
	}
	return otp, nil
}

// IsAuthenticated checks if a request is authenticated
func (s *Service) IsAuthenticated(r *http.Request) bool {
	_, err := GetUserFromContext(r.Context())
	return err == nil
}

// AddUserToContext adds user claims to a context
func AddUserToContext(ctx context.Context, claims *TokenClaims) context.Context {
	return context.WithValue(ctx, UserContextKey, claims)
}

// GetUserFromContext extracts user claims from request context
func GetUserFromContext(ctx context.Context) (*TokenClaims, error) {
	user, ok := ctx.Value(UserContextKey).(*TokenClaims)
	if !ok {
		return nil, ErrUserNotInContext
	}
	return user, nil
}

// ValidateEmail checks if an email is well-formed
func ValidateEmail(email string) bool {
	// Use validator to check email
	validate := validator.New()
	err := validate.Var(email, "required,email")
	return err == nil
}

// SanitizeUsername removes potentially harmful characters from username
func SanitizeUsername(username string) string {
	// This is a simple implementation
	// In a production system, you might want more sophisticated sanitization
	sanitized := strings.TrimSpace(username)
	// Remove any characters that aren't alphanumeric, underscore, or period
	sanitized = regexp.MustCompile(`[^a-zA-Z0-9_.]+`).ReplaceAllString(sanitized, "")
	return sanitized
}