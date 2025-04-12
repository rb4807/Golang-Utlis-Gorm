package router

import (
	"fmt"
	"net/http"

	"github.com/rb4807/Golang-Utlis-Postgresql/auth"
	"github.com/rb4807/Golang-Utlis-Postgresql/controller"
)

func UserRoutes(mux *http.ServeMux, authService *auth.Service) {
	baseAppPath := "/api/user"

	// Public

	// Protected
	mux.Handle(fmt.Sprintf("%s/get_user_profile", baseAppPath), authService.AuthMiddleware(controller.GetUserProfile(authService)))
	mux.Handle(fmt.Sprintf("%s/change_user_password", baseAppPath), authService.AuthMiddleware(controller.ChangeUserPassword(authService)))	
}