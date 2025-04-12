package controller

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/rb4807/Golang-Utlis-Postgresql/auth"
	"github.com/rb4807/Golang-Utlis-Postgresql/dto"
	"github.com/rb4807/Golang-Utlis-Postgresql/middleware"
	"github.com/rb4807/Golang-Utlis-Postgresql/utils"
)

func UserRegister(authService *auth.Service) http.HandlerFunc {
	handler := func(w http.ResponseWriter, r *http.Request) {
		var req dto.RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.SendJSONError(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		user := auth.User{
			Username:  req.Email,
			Email:     req.Email,
			FirstName: req.FirstName,
			LastName:  req.LastName,
			IsActive:  true,
		}

		userID, err := authService.Register(user, req.Password)
		if err != nil {
			switch {
			case err == auth.ErrEmailExists:
				utils.SendJSONError(w, "Email already exists", http.StatusBadRequest)
			case err == auth.ErrUsernameExists:
				utils.SendJSONError(w, "Username already taken", http.StatusBadRequest)
			default:
				utils.SendJSONError(w, err.Error(), http.StatusBadRequest)
			}
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":         userID,
			"email":      user.Email,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"username":   user.Username,
			"message":    "User registered successfully",
		})
	}

	return middleware.RequestMethodValidator([]string{http.MethodPost}, handler)
}

func UserLogin(authService *auth.Service) http.HandlerFunc {
	handler := func(w http.ResponseWriter, r *http.Request) {
		var req dto.LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.SendJSONError(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		user, token, err := authService.Login(req.Username, req.Password)
		if err != nil {
			switch err {
			case auth.ErrUserNotFound:
				utils.SendJSONError(w, "No account found with these details", http.StatusUnauthorized)
			// case auth.ErrUserInactive:
			//     utils.SendJSONError(w, "Account is inactive. Please contact support.", http.StatusForbidden)
			case auth.ErrInvalidPassword:
				utils.SendJSONError(w, "Invalid password", http.StatusUnauthorized)
			default:
				utils.SendJSONError(w, "Login failed", http.StatusInternalServerError)
			}
			return
		}

		expiresAt := time.Now().Add(24 * time.Hour)
		response := dto.TokenResponse{
			Token:     token,
			ExpiresAt: expiresAt,
			UserID:    uint64(user.ID),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}

	return middleware.RequestMethodValidator([]string{http.MethodPost}, handler)
}

func AdminHandler(w http.ResponseWriter, r *http.Request) {
	claims, _ := auth.GetUserFromContext(r.Context())

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "Welcome to the admin area",
		"user_id":  claims.UserID,
		"username": claims.Username,
	})
}

func SuperuserHandler(w http.ResponseWriter, r *http.Request) {
	claims, _ := auth.GetUserFromContext(r.Context())

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "Welcome to the superuser area",
		"user_id":  claims.UserID,
		"username": claims.Username,
	})
}
