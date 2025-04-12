package router

import (
	"fmt"
	"net/http"

	"github.com/rb4807/Golang-Utlis-Postgresql/auth"
	"github.com/rb4807/Golang-Utlis-Postgresql/controller"
)

func AuthRoutes(mux *http.ServeMux, authService *auth.Service) {
	baseAppPath := "/api/auth"

	// Public
	mux.HandleFunc(fmt.Sprintf("%s/user_register", baseAppPath), controller.UserRegister(authService))
	mux.HandleFunc(fmt.Sprintf("%s/user_login", baseAppPath), controller.UserLogin(authService))

	// Protected
	mux.Handle(fmt.Sprintf("%s/admin", baseAppPath), authService.AdminMiddleware(http.HandlerFunc(controller.AdminHandler)))
	mux.Handle(fmt.Sprintf("%s/superuser", baseAppPath), authService.SuperuserMiddleware(http.HandlerFunc(controller.SuperuserHandler)))
}

