package httpapi

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"github.com/grachevelloper/frets/services/internal/platform"
	"github.com/grachevelloper/frets/services/song-library/internal/library"
)

type SessionValidator interface {
	Validate(context.Context, string) (string, error)
}

type Handler struct {
	library  *library.Service
	sessions SessionValidator
}

func Register(mux *http.ServeMux, service *library.Service, sessions SessionValidator) {
	h := Handler{library: service, sessions: sessions}
	mux.HandleFunc("GET /api/v1/songs", h.authenticated(h.list))
	mux.HandleFunc("POST /api/v1/songs", h.authenticated(h.upload))
	mux.HandleFunc("GET /api/v1/songs/{songId}", h.authenticated(h.detail))
	mux.HandleFunc("GET /api/v1/songs/{songId}/file", h.authenticated(h.download))
	mux.HandleFunc("PUT /api/v1/songs/{songId}/fragments/{fragmentId}", h.authenticated(h.putFragment))
	mux.HandleFunc("DELETE /api/v1/songs/{songId}", h.authenticated(h.delete))
}

func (h Handler) authenticated(next func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(platform.SessionCookie)
		if err != nil || cookie.Value == "" {
			platform.Error(w, 401, "unauthorized")
			return
		}

		owner, err := h.sessions.Validate(r.Context(), cookie.Value)
		if errors.Is(err, platform.ErrUnauthorized) {
			platform.Error(w, 401, "unauthorized")
			return
		}
		if err != nil {
			platform.Unavailable(w, err)
			return
		}

		next(w, r, owner)
	}
}

func validID(id string) bool {
	parsed, err := uuid.Parse(id)
	return err == nil && parsed.String() == id
}

func songID(w http.ResponseWriter, r *http.Request) (string, bool) {
	id := r.PathValue("songId")
	if !validID(id) {
		platform.Error(w, 404, "not_found")
		return "", false
	}
	return id, true
}

func failure(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, library.ErrNotFound):
		platform.Error(w, 404, "not_found")
	case errors.Is(err, library.ErrInvalidUpload), errors.Is(err, library.ErrInvalidTitle),
		errors.Is(err, library.ErrChecksum), errors.Is(err, library.ErrInvalidFragment):
		platform.Error(w, 400, err.Error())
	case errors.Is(err, library.ErrFileTooLarge):
		platform.Error(w, 413, "file_too_large")
	case errors.Is(err, library.ErrUploadConflict), errors.Is(err, library.ErrUploadInProgress),
		errors.Is(err, library.ErrUploadExpired), errors.Is(err, library.ErrFragmentConflict):
		platform.Error(w, 409, err.Error())
	default:
		platform.Unavailable(w, err)
	}
}
