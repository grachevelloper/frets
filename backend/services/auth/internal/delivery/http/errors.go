package httpapi

import (
	"crypto/subtle"
	"errors"
	"log/slog"
	"net/http"

	"github.com/grachevelloper/frets/backend/lib/httpjson"
	"github.com/grachevelloper/frets/backend/services/auth/internal/domain/auth"
)

func validServiceSecret(header, secret string) bool {
	return subtle.ConstantTimeCompare(auth.HashSecret(header), auth.HashSecret("Bearer "+secret)) == 1
}

func failure(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, auth.ErrUnauthorized):
		httpjson.Error(w, http.StatusUnauthorized, "unauthorized")
	case errors.Is(err, auth.ErrUnknownProvider):
		httpjson.Error(w, http.StatusNotFound, "not_found")
	case errors.Is(err, auth.ErrProviderUnavailable):
		httpjson.Error(w, http.StatusServiceUnavailable, "provider_unavailable")
	case errors.Is(err, auth.ErrLoginInvalid):
		httpjson.Error(w, http.StatusBadRequest, "login_invalid")
	case errors.Is(err, auth.ErrLoginFailed):
		slog.Error("OAuth login failed", "error", err)
		httpjson.Error(w, http.StatusBadGateway, "login_failed")
	default:
		slog.Error("auth request failed", "error", err)
		httpjson.Error(w, http.StatusServiceUnavailable, "unavailable")
	}
}

func NotFound(w http.ResponseWriter, r *http.Request) {
	httpjson.Error(w, http.StatusNotFound, "not_found")
}
