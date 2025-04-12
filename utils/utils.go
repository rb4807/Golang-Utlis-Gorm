package utils

import (
	"encoding/json"
	"net/http"
)

func SendJSONError(w http.ResponseWriter, message string, statusCode int, errDetail ...string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := map[string]string{
		"message": message,
	}

	if len(errDetail) > 0 && errDetail[0] != "" {
		response["error"] = errDetail[0]
	}

	json.NewEncoder(w).Encode(response)
}
