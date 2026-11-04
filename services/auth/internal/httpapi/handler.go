package httpapi

import (
	"errors"
	"net/http"
	"time"

	"github.com/grachevelloper/frets/services/auth/internal/account"
	"github.com/grachevelloper/frets/services/internal/platform"
)

const browserCookie = "frets_login"

type Config struct {
	InternalSecret string
	SecureCookies  bool
	LocalLogin     bool
}

type Handler struct {
	sessions *account.SessionService
	login    *account.LoginService
	config   Config
}

func Register(mux *http.ServeMux, sessions *account.SessionService, login *account.LoginService, config Config) {
	h := Handler{sessions: sessions, login: login, config: config}
	mux.HandleFunc("GET /api/v1/auth/login/{provider}", h.beginLogin)
	mux.HandleFunc("GET /api/v1/auth/callback/{provider}", h.callback)
	mux.HandleFunc("GET /api/v1/me", h.me)
	mux.HandleFunc("POST /api/v1/auth/logout", h.logout)
	mux.HandleFunc("POST /internal/sessions/validate", h.validate)
	if config.LocalLogin {
		mux.HandleFunc("POST /api/v1/auth/local", h.localLogin)
	}
}

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
	h.cookie(w, platform.SessionCookie, key, int(account.SessionLifetime/time.Second))
}

func failure(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, account.ErrUnauthorized):
		platform.Error(w, 401, "unauthorized")
	case errors.Is(err, account.ErrUnknownProvider):
		platform.Error(w, 404, "not_found")
	case errors.Is(err, account.ErrProviderUnavailable):
		platform.Error(w, 503, "provider_unavailable")
	case errors.Is(err, account.ErrLoginInvalid):
		platform.Error(w, 400, "login_invalid")
	case errors.Is(err, account.ErrLoginFailed):
		platform.Error(w, 502, "login_failed")
	default:
		platform.Unavailable(w, err)
	}
}
