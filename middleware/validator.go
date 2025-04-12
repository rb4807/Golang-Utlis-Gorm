package middleware

import (
	"net/http"
	"github.com/rb4807/Golang-Utlis-Postgresql/utils"
)

func RequestMethodValidator(allowedMethods []string, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		for _, method := range allowedMethods {
			if r.Method == method {
				handler(w, r)
				return
			}
		}
		utils.SendJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}