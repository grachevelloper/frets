package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/grachevelloper/frets/backend/lib/httpjson"
	"github.com/grachevelloper/frets/backend/services/song-library/internal/domain/song"
)

type SessionValidator interface {
	Validate(ctx context.Context, sessionToken string) (userID string, err error)
}

type AuthenticatedHandler func(http.ResponseWriter, *http.Request, string)

func Session(sessions SessionValidator, next AuthenticatedHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("frets_session")
		if err != nil || cookie.Value == "" {
			httpjson.Error(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		owner, err := sessions.Validate(r.Context(), cookie.Value)
		if errors.Is(err, song.ErrUnauthorized) {
			httpjson.Error(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		if err != nil {
			slog.Error("validate session", "error", err)
			httpjson.Error(w, http.StatusServiceUnavailable, "unavailable")
			return
		}
		next(w, r, owner)
	}
}
