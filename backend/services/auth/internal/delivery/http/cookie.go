package httpapi

import (
	"net/http"
	"time"

	"github.com/grachevelloper/frets/backend/services/auth/internal/domain/auth"
)

const (
	browserCookie     = "frets_login"
	sessionCookieName = "frets_session"
)

func (h Handler) cookie(w http.ResponseWriter, name, value string, maxAge int) {
	http.SetCookie(w, &http.Cookie{
		Name: name, Value: value, Path: "/", MaxAge: maxAge,
		HttpOnly: true, Secure: h.config.SecureCookies, SameSite: http.SameSiteLaxMode,
	})
}

func cookieValue(r *http.Request, name string) string {
	cookie, err := r.Cookie(name)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func (h Handler) sessionCookie(w http.ResponseWriter, key string) {
	h.cookie(w, sessionCookieName, key, int(auth.SessionLifetime/time.Second))
}
