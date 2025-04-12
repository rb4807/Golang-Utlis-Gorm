package router

import (
	"net/http"

	"github.com/rb4807/Golang-Utlis-Postgresql/auth"
	"github.com/rb4807/Golang-Utlis-Postgresql/middleware"
)

func InitRoutes(authService *auth.Service) http.Handler {
	mux := http.NewServeMux()

	UserRoutes(mux, authService)
	AuthRoutes(mux, authService)

	return middleware.LoggingMiddleware(middleware.ErrorCatchMiddleware(middleware.PageNotFoundMiddleware(mux)))
}
