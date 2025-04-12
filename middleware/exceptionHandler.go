package middleware

import (
	"fmt"
	"log"
	"net/http"
	"github.com/rb4807/Golang-Utlis-Postgresql/utils"
)

func PageNotFoundMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if mux, ok := next.(*http.ServeMux); ok {
			_, pattern := mux.Handler(r)
			if pattern == "" {
				utils.SendJSONError(w, "Page not found", http.StatusNotFound)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func ErrorCatchMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Recovered from panic: %v", err)

				utils.SendJSONError(w, "Internal server error", http.StatusInternalServerError, fmt.Sprintf("%v", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}