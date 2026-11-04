package httpapi

import (
	"errors"
	"net/http"

	"github.com/grachevelloper/frets/services/auth/internal/account"
	"github.com/grachevelloper/frets/services/internal/platform"
)

func (h Handler) me(w http.ResponseWriter, r *http.Request) {
	profile, err := h.sessions.Profile(r.Context(), cookieValue(r, platform.SessionCookie))
	if err != nil {
		failure(w, err)
		return
	}
	platform.JSON(w, 200, profile)
}

func (h Handler) validate(w http.ResponseWriter, r *http.Request) {
	if !platform.EqualSecret(r.Header.Get("Authorization"), "Bearer "+h.config.InternalSecret) {
		platform.Error(w, 403, "forbidden")
		return
	}

	profile, err := h.sessions.Profile(r.Context(), cookieValue(r, platform.SessionCookie))
	if err != nil {
		failure(w, err)
		return
	}
	platform.JSON(w, 200, map[string]string{"userId": profile.ID})
}

func (h Handler) logout(w http.ResponseWriter, r *http.Request) {
	err := h.sessions.Revoke(r.Context(), cookieValue(r, platform.SessionCookie))
	if err == nil || errors.Is(err, account.ErrUnauthorized) {
		h.cookie(w, platform.SessionCookie, "", -1)
	}
	if err != nil {
		failure(w, err)
		return
	}
	w.WriteHeader(204)
}

func (h Handler) localLogin(w http.ResponseWriter, r *http.Request) {
	key, err := h.sessions.Create(r.Context(), "local", account.Identity{ID: "local", Name: "Local guitarist"})
	if err != nil {
		failure(w, err)
		return
	}
	h.sessionCookie(w, key)
	w.WriteHeader(204)
}
