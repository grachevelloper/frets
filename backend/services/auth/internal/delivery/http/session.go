package httpapi

import (
	"errors"
	"net/http"

	"github.com/grachevelloper/frets/backend/lib/httpjson"
	"github.com/grachevelloper/frets/backend/services/auth/internal/domain/auth"
)

func (h Handler) Me(w http.ResponseWriter, r *http.Request) {
	profile, err := h.sessions.Profile(r.Context(), cookieValue(r, sessionCookieName))
	if err != nil {
		failure(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, profileResponse(profile))
}

func (h Handler) Validate(w http.ResponseWriter, r *http.Request) {
	if !validServiceSecret(r.Header.Get("Authorization"), h.config.InternalSecret) {
		httpjson.Error(w, http.StatusForbidden, "forbidden")
		return
	}

	profile, err := h.sessions.Profile(r.Context(), cookieValue(r, sessionCookieName))
	if err != nil {
		failure(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, map[string]string{"userId": profile.ID})
}

func (h Handler) Logout(w http.ResponseWriter, r *http.Request) {
	err := h.sessions.Revoke(r.Context(), cookieValue(r, sessionCookieName))
	if err == nil || errors.Is(err, auth.ErrUnauthorized) {
		h.cookie(w, sessionCookieName, "", -1)
	}
	if err != nil {
		failure(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h Handler) LocalLogin(w http.ResponseWriter, r *http.Request) {
	key, err := h.sessions.Create(r.Context(), "local", auth.Identity{
		ID:   "local",
		Name: "Local guitarist",
	})
	if err != nil {
		failure(w, err)
		return
	}
	h.sessionCookie(w, key)
	w.WriteHeader(http.StatusNoContent)
}
