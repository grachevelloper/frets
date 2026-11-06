package delivery

import (
	"net/http"

	"github.com/grachevelloper/frets/backend/lib/httpmiddleware"
	httpapi "github.com/grachevelloper/frets/backend/services/song-library/internal/delivery/http"
	"github.com/grachevelloper/frets/backend/services/song-library/internal/delivery/middleware"
)

func NewRouter(origin string, handler *httpapi.Handler, sessions middleware.SessionValidator) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/songs", middleware.Session(sessions, handler.List))
	mux.HandleFunc("POST /api/v1/songs", middleware.Session(sessions, handler.Upload))
	mux.HandleFunc("GET /api/v1/songs/{songId}", middleware.Session(sessions, handler.GetSong))
	mux.HandleFunc("GET /api/v1/songs/{songId}/file", middleware.Session(sessions, handler.Download))
	mux.HandleFunc("PUT /api/v1/songs/{songId}/fragments/{fragmentId}", middleware.Session(sessions, handler.CreateFragment))
	mux.HandleFunc("DELETE /api/v1/songs/{songId}", middleware.Session(sessions, handler.Delete))
	mux.HandleFunc("/", httpapi.NotFound)
	return httpmiddleware.Origin(origin, mux)
}
