package httpapi

import (
	"net/http"
	"time"

	"github.com/grachevelloper/frets/backend/services/auth/internal/domain/auth"
)

func (h Handler) BeginLogin(w http.ResponseWriter, r *http.Request) {
	location, browser, err := h.login.Begin(r.Context(), r.PathValue("provider"), r.URL.Query().Get("next"))
	if err != nil {
		failure(w, err)
		return
	}

	h.cookie(w, browserCookie, browser, int(auth.LoginLifetime/time.Second))
	http.Redirect(w, r, location, http.StatusFound)
}

func (h Handler) Callback(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	key, next, consumed, err := h.login.Complete(
		r.Context(), r.PathValue("provider"), cookieValue(r, browserCookie), query.Get("state"), query.Get("code"),
	)
	if consumed {
		h.cookie(w, browserCookie, "", -1)
	}
	if err != nil {
		failure(w, err)
		return
	}

	h.sessionCookie(w, key)
	http.Redirect(w, r, next, http.StatusFound)
}
