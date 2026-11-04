package main

import (
	"net/http"
	"os"
	"strings"

	"github.com/grachevelloper/frets/services/auth/internal/account"
	"github.com/grachevelloper/frets/services/auth/internal/httpapi"
	"github.com/grachevelloper/frets/services/auth/internal/oauth"
	"github.com/grachevelloper/frets/services/auth/internal/postgres"
	"github.com/grachevelloper/frets/services/internal/platform"
)

func main() {
	origin := platform.Required("PUBLIC_ORIGIN")
	db := platform.Database("auth", postgres.Schema)
	repo := postgres.New(db)
	sessions := account.NewSessions(repo)
	login := account.NewLogin(repo, oauth.FromEnvironment(origin), sessions)

	mux := http.NewServeMux()
	httpapi.Register(mux, sessions, login, httpapi.Config{
		InternalSecret: platform.Required("INTERNAL_SECRET"),
		SecureCookies:  strings.HasPrefix(origin, "https://"),
		LocalLogin:     origin == "http://localhost:6666" && os.Getenv("LOCAL_LOGIN") == "true",
	})
	platform.Serve(mux)
}
