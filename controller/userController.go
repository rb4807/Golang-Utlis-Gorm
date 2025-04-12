package controller

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/rb4807/Golang-Utlis-Postgresql/auth"
	"github.com/rb4807/Golang-Utlis-Postgresql/dto"
	"github.com/rb4807/Golang-Utlis-Postgresql/middleware"
	"github.com/rb4807/Golang-Utlis-Postgresql/utils"
)

func GetUserProfile(authService *auth.Service) http.HandlerFunc {
	handler := func(w http.ResponseWriter, r *http.Request) {
		claims, err := auth.GetUserFromContext(r.Context())
		if err != nil {
			utils.SendJSONError(w, "User not found", http.StatusUnauthorized)
			return
		}

		user, err := authService.GetUserByID(claims.UserID)
		if err != nil {
			utils.SendJSONError(w, "Error retrieving user", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"user_id":      user.ID,
			"username":     user.Username,
			"email":        user.Email,
			"first_name":   user.FirstName,
			"last_name":    user.LastName,
			"is_superuser": user.IsSuperuser,
			"date_joined":  user.DateJoined,
			"last_login":   user.LastLogin,
		})
	}

	return middleware.RequestMethodValidator([]string{http.MethodGet}, handler)
}

func ChangeUserPassword(authService *auth.Service) http.HandlerFunc {
	handler := func(w http.ResponseWriter, r *http.Request) {
		claims, err := auth.GetUserFromContext(r.Context())
		if err != nil {
			utils.SendJSONError(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		userID := claims.UserID

		var req dto.ChangePasswordRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.SendJSONError(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.CurrentPassword == "" || req.NewPassword == "" {
			utils.SendJSONError(w, "Current and new passwords are required", http.StatusBadRequest)
			return
		}

		err = authService.ChangePassword(userID, req.CurrentPassword, req.NewPassword)
		if err != nil {
			switch {
			case errors.Is(err, auth.ErrInvalidPassword):
				utils.SendJSONError(w, "Current password is incorrect", http.StatusUnauthorized)
			case errors.Is(err, auth.ErrUserNotFound):
				utils.SendJSONError(w, "User not found", http.StatusNotFound)
			default:
				utils.SendJSONError(w, "Failed to change password", http.StatusInternalServerError)
			}
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Password changed successfully"})
	}

	return middleware.RequestMethodValidator([]string{http.MethodPost}, handler)
}