package httpjson

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type ErrorResponse struct {
	Code string `json:"code"`
}

func Write(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		slog.Error("write JSON response", "error", err)
	}
}

func Error(w http.ResponseWriter, status int, code string) {
	Write(w, status, ErrorResponse{Code: code})
}
