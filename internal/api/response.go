package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type Response struct {
	Data  any    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

func SendJSON(w http.ResponseWriter, resp Response, status int){
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(resp) ; err != nil {
		slog.Error("failed to encode json", "error", err)
	}
}