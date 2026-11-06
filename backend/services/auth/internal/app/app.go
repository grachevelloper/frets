package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/grachevelloper/frets/backend/lib/httpserver"
	db "github.com/grachevelloper/frets/backend/lib/postgres"
	"github.com/grachevelloper/frets/backend/services/auth/internal/config"
	"github.com/grachevelloper/frets/backend/services/auth/internal/delivery"
	httpapi "github.com/grachevelloper/frets/backend/services/auth/internal/delivery/http"
	"github.com/grachevelloper/frets/backend/services/auth/internal/infrastructure/oauth"
	"github.com/grachevelloper/frets/backend/services/auth/internal/infrastructure/postgres"
	"github.com/grachevelloper/frets/backend/services/auth/internal/infrastructure/token"
	"github.com/grachevelloper/frets/backend/services/auth/internal/usecase/login"
	"github.com/grachevelloper/frets/backend/services/auth/internal/usecase/session"
	"github.com/grachevelloper/frets/backend/services/auth/migrations"
)

type App struct {
	db      *pgxpool.Pool
	address string
	handler http.Handler
}

func New(ctx context.Context) (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	startupCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	pool, err := db.Open(startupCtx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}
	if err := db.Migrate(startupCtx, pool, "auth", migrations.Files); err != nil {
		pool.Close()
		return nil, fmt.Errorf("migrate auth: %w", err)
	}
	repo := postgres.New(pool)
	tokens := token.Generator{}
	sessions := session.New(repo, tokens)

	providers := map[string]login.Provider{
		"google": nil,
		"yandex": nil,
	}
	if cfg.Google.ClientID != "" && cfg.Google.ClientSecret != "" {
		providers["google"] = oauth.NewGoogle(cfg.PublicOrigin, oauth.Credentials{
			ClientID:     cfg.Google.ClientID,
			ClientSecret: cfg.Google.ClientSecret,
		})
	} else {
		missing := []string{}
		if cfg.Google.ClientID == "" {
			missing = append(missing, "GOOGLE_CLIENT_ID")
		}
		if cfg.Google.ClientSecret == "" {
			missing = append(missing, "GOOGLE_CLIENT_SECRET")
		}
		slog.Warn("OAuth provider is disabled", "provider", "google", "missingSettings", missing)
	}
	if cfg.Yandex.ClientID != "" && cfg.Yandex.ClientSecret != "" {
		providers["yandex"] = oauth.NewYandex(cfg.PublicOrigin, oauth.Credentials{
			ClientID:     cfg.Yandex.ClientID,
			ClientSecret: cfg.Yandex.ClientSecret,
		})
	} else {
		missing := []string{}
		if cfg.Yandex.ClientID == "" {
			missing = append(missing, "YANDEX_CLIENT_ID")
		}
		if cfg.Yandex.ClientSecret == "" {
			missing = append(missing, "YANDEX_CLIENT_SECRET")
		}
		slog.Warn("OAuth provider is disabled", "provider", "yandex", "missingSettings", missing)
	}

	logins := login.New(repo, providers, sessions, tokens)
	handler := httpapi.New(sessions, logins, httpapi.CookieConfig{
		InternalSecret: cfg.InternalSecret,
		SecureCookies:  strings.HasPrefix(cfg.PublicOrigin, "https://"),
	})
	return &App{
		db:      pool,
		address: cfg.Address,
		handler: delivery.NewRouter(cfg.PublicOrigin, handler, cfg.LocalLogin),
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	return httpserver.Run(ctx, a.address, a.handler)
}

func (a *App) Close() {
	a.db.Close()
}
