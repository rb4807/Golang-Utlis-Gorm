package auth

import (
	"errors"
	"time"
	
	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	Username        string    `gorm:"size:50;uniqueIndex" json:"username" validate:"required,min=3,max=50"`
	Email           string    `gorm:"size:100;uniqueIndex" json:"email" validate:"required,email"`
	Password        string    `gorm:"size:255" json:"-"` // Hashed password, never expose in JSON
	FirstName       string    `gorm:"size:50" json:"first_name"`
	LastName        string    `gorm:"size:50" json:"last_name"`
	IsActive        bool      `gorm:"default:true" json:"is_active"`
	IsSuperuser     bool      `gorm:"default:false" json:"is_superuser"`
	DateJoined      time.Time `gorm:"autoCreateTime" json:"date_joined"`
	LastLogin       *time.Time `json:"last_login"`
	PasswordChanged *time.Time `json:"password_changed"`
}

// OTP stores one-time password information
type OTP struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"index"`
	User      User      `gorm:"constraint:OnDelete:CASCADE;"`
	OTPValue  string    `gorm:"size:10;column:otp_value" json:"otp_value"`
	ExpiresAt time.Time `json:"expires_at"`
	Verified  bool      `gorm:"default:false" json:"verified"`
}

// Config holds the configuration for the authentication package
type Config struct {
	JWTSecret     string
	TokenDuration time.Duration
	DB            *gorm.DB
}

// Service provides authentication functionality
type Service struct {
	config    Config
	validator interface{} // This will be a *validator.Validate
}

// Common errors
var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidPassword    = errors.New("current password is incorrect")
	ErrUserNotInContext   = errors.New("user not found in context")
	ErrConfigInvalid      = errors.New("configuration is invalid")
	ErrEmailExists        = errors.New("email already exists")
	ErrUsernameExists     = errors.New("username already exists")
)

// Initialize database tables
func InitDB(db *gorm.DB) error {
	// Auto migrate will create or modify tables based on struct definitions
	return db.AutoMigrate(&User{}, &OTP{})
}

// GetUserByID retrieves a user by ID
func (s *Service) GetUserByID(userID uint) (*User, error) {
	var user User
	result := s.config.DB.First(&user, userID)
	
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, result.Error
	}
	
	return &user, nil
}

// UpdateUser updates user information
func (s *Service) UpdateUser(user *User) error {
	// Only update specific fields, not the entire record
	result := s.config.DB.Model(user).Updates(map[string]interface{}{
		"username":     user.Username,
		"email":        user.Email,
		"first_name":   user.FirstName,
		"last_name":    user.LastName,
		"is_active":    user.IsActive,
		"is_superuser": user.IsSuperuser,
	})
	
	return result.Error
}