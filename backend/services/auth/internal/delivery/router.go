package delivery

import (
	"net/http"

	"github.com/grachevelloper/frets/backend/lib/httpmiddleware"
	httpapi "github.com/grachevelloper/frets/backend/services/auth/internal/delivery/http"
)

func NewRouter(origin string, handler *httpapi.Handler, localLogin bool) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/auth/login/{provider}", handler.BeginLogin)
	mux.HandleFunc("GET /api/v1/auth/callback/{provider}", handler.Callback)
	mux.HandleFunc("GET /api/v1/me", handler.Me)
	mux.HandleFunc("POST /api/v1/auth/logout", handler.Logout)
	if localLogin {
		mux.HandleFunc("POST /api/v1/auth/local", handler.LocalLogin)
	}
	mux.HandleFunc("/", httpapi.NotFound)
	router := http.NewServeMux()
	router.HandleFunc("POST /internal/sessions/validate", handler.Validate)
	router.Handle("/", httpmiddleware.Origin(origin, mux))
	return router
}
