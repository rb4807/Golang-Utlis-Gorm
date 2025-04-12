package auth

import (
	"errors"
	"time"
	"strings"
	"gorm.io/gorm"
	"log"
)

func (s *Service) Register(user User, password string) (uint, error) {
    // Use email as username if needed
    if user.Username == "" {
        user.Username = user.Email
    }

    // Validate user input
    if err := s.validateData(user); err != nil {
        return 0, err
    }

    // Hash the password
    hashedPassword, err := s.HashPassword(password)
    if err != nil {
        return 0, err
    }
    
    user.Password = hashedPassword
    user.DateJoined = time.Now()
    now := time.Now()
    user.PasswordChanged = &now

    // Create the user
    result := s.config.DB.Create(&user)
    if result.Error != nil {
        // Check for unique constraint violations
        if strings.Contains(result.Error.Error(), "duplicate key value violates unique constraint") {
            if strings.Contains(result.Error.Error(), "idx_users_email") {
                return 0, ErrEmailExists
            }
            if strings.Contains(result.Error.Error(), "idx_users_username") {
                return 0, ErrUsernameExists
            }
        }
        return 0, result.Error
    }

    return user.ID, nil
}
// Authenticate verifies a user's credentials
func (s *Service) Authenticate(username, password string) (*User, error) {
    var user User
    
    // First check if any user exists with this username/email (active or inactive)
    result := s.config.DB.Where("username = ? OR email = ?", username, username).First(&user)
    if result.Error != nil {
        if errors.Is(result.Error, gorm.ErrRecordNotFound) {
            return nil, ErrUserNotFound // New error type for this case
        }
        return nil, result.Error
    }

    // // Then check if the user is active
    // if !user.IsActive {
    //     return nil, ErrUserInactive // New error type for inactive users
    // }

    // Finally verify the password
    if !s.VerifyPassword(user.Password, password) {
        return nil, ErrInvalidPassword // More specific than ErrInvalidCredentials
    }

    // Update last login time
    now := time.Now()
    user.LastLogin = &now
    if err := s.config.DB.Model(&user).Update("last_login", now).Error; err != nil {
        // Log this error but don't fail authentication because of it
        log.Printf("Failed to update last login time: %v", err)
    }

    return &user, nil
}

// Login combines authentication and JWT generation
func (s *Service) Login(username, password string) (*User, string, error) {
	user, err := s.Authenticate(username, password)
	if err != nil {
		return nil, "", err
	}

	token, err := s.GenerateJWT(user)
	if err != nil {
		return user, "", err
	}

	return user, token, nil
}

// GenerateOTP creates a one-time password for a user
func (s *Service) GenerateOTP(userID uint, length int, validityMinutes int) (string, error) {
	if length <= 0 {
		length = 6 // Default OTP length
	}
	if validityMinutes <= 0 {
		validityMinutes = 15 // Default validity: 15 minutes
	}

	// Check if user exists
	var user User
	result := s.config.DB.First(&user, userID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return "", ErrUserNotFound
		}
		return "", result.Error
	}

	// Generate random OTP
	otpValue, err := s.generateRandomOTP(length)
	if err != nil {
		return "", err
	}

	// Delete any existing OTPs for this user
	s.config.DB.Where("user_id = ?", userID).Delete(&OTP{})

	// Create new OTP
	otp := OTP{
		UserID:    userID,
		OTPValue:  otpValue,
		ExpiresAt: time.Now().Add(time.Duration(validityMinutes) * time.Minute),
		Verified:  false,
	}
	
	result = s.config.DB.Create(&otp)
	if result.Error != nil {
		return "", result.Error
	}

	return otpValue, nil
}

// VerifyOTP checks if an OTP is valid for a user
func (s *Service) VerifyOTP(userID uint, otpValue string) (bool, error) {
	var otp OTP
	
	result := s.config.DB.Where(
		"user_id = ? AND otp_value = ? AND expires_at > ? AND verified = ?", 
		userID, otpValue, time.Now(), false,
	).First(&otp)
	
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, result.Error
	}

	// Mark OTP as verified
	otp.Verified = true
	s.config.DB.Save(&otp)

	return true, nil
}

// ChangePassword updates a user's password
func (s *Service) ChangePassword(userID uint, currentPassword, newPassword string) error {
	// Get current user details
	var user User
	result := s.config.DB.First(&user, userID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return result.Error
	}

	// Verify current password
	if !s.VerifyPassword(user.Password, currentPassword) {
		return ErrInvalidPassword
	}

	// Hash new password
	hashedPassword, err := s.HashPassword(newPassword)
	if err != nil {
		return err
	}

	// Update password
	now := time.Now()
	updates := map[string]interface{}{
		"password":         hashedPassword,
		"password_changed": now,
	}
	
	return s.config.DB.Model(&user).Updates(updates).Error
}

// ResetPassword resets a user's password (admin function or after verification)
func (s *Service) ResetPassword(userID uint, newPassword string) error {
	// Check if user exists
	var user User
	result := s.config.DB.First(&user, userID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return result.Error
	}

	// Hash new password
	hashedPassword, err := s.HashPassword(newPassword)
	if err != nil {
		return err
	}

	// Update password
	now := time.Now()
	updates := map[string]interface{}{
		"password":         hashedPassword,
		"password_changed": now,
	}
	
	return s.config.DB.Model(&user).Updates(updates).Error
}

// UserExists checks if a user exists by ID and/or username
func (s *Service) UserExists(userID uint, username string) (bool, error) {
	var count int64
	
	if userID > 0 && username != "" {
		s.config.DB.Model(&User{}).Where("id = ? AND username = ?", userID, username).Count(&count)
	} else if userID > 0 {
		s.config.DB.Model(&User{}).Where("id = ?", userID).Count(&count)
	} else if username != "" {
		s.config.DB.Model(&User{}).Where("username = ?", username).Count(&count)
	} else {
		return false, errors.New("at least one of userID or username must be provided")
	}
	
	return count > 0, nil
}