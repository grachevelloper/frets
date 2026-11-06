package httpapi

import (
	"context"

	"github.com/grachevelloper/frets/backend/services/auth/internal/domain/auth"
)

type Sessions interface {
	Create(ctx context.Context, provider string, identity auth.Identity) (sessionToken string, err error)
	Profile(ctx context.Context, sessionToken string) (profile auth.Profile, err error)
	Revoke(ctx context.Context, sessionToken string) (err error)
}

type Login interface {
	Begin(ctx context.Context, provider string, returnTo string) (location string, browserToken string, err error)
	Complete(ctx context.Context, provider string, browserToken string, state string, code string) (sessionToken string, returnTo string, consumed bool, err error)
}

type CookieConfig struct {
	InternalSecret string
	SecureCookies  bool
}

type Handler struct {
	sessions Sessions
	login    Login
	config   CookieConfig
}

func New(sessions Sessions, login Login, config CookieConfig) *Handler {
	return &Handler{
		sessions: sessions,
		login:    login,
		config:   config,
	}
}
