package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/grachevelloper/frets/services/internal/platform"
)

//go:embed schema.sql
var schema string

type server struct {
	db              *pgxpool.Pool
	files           storage
	authURL, secret string
	authClient      *http.Client
}

func main() {
	s := server{
		db:         platform.Database("song_library", schema),
		files:      newStorage(),
		authURL:    platform.Required("AUTH_URL") + "/internal/sessions/validate",
		secret:     platform.Required("INTERNAL_SECRET"),
		authClient: &http.Client{Timeout: 3 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }},
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/songs", s.authenticated(s.list))
	mux.HandleFunc("POST /api/v1/songs", s.authenticated(s.upload))
	mux.HandleFunc("GET /api/v1/songs/{songId}", s.authenticated(s.detail))
	mux.HandleFunc("GET /api/v1/songs/{songId}/file", s.authenticated(s.download))
	mux.HandleFunc("PUT /api/v1/songs/{songId}/fragments/{fragmentId}", s.authenticated(s.putFragment))
	mux.HandleFunc("DELETE /api/v1/songs/{songId}", s.authenticated(s.delete))
	go s.maintenance()
	platform.Serve(mux)
}

func (s server) authenticated(next func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(platform.SessionCookie)
		if err != nil || cookie.Value == "" {
			platform.Error(w, 401, "unauthorized")
			return
		}

		req, err := http.NewRequestWithContext(r.Context(), "POST", s.authURL, nil)
		if err != nil {
			platform.Unavailable(w, err)
			return
		}
		req.AddCookie(cookie)
		req.Header.Set("Authorization", "Bearer "+s.secret)
		response, err := s.authClient.Do(req)
		if err != nil {
			platform.Unavailable(w, err)
			return
		}
		defer response.Body.Close()
		if response.StatusCode == 401 {
			platform.Error(w, 401, "unauthorized")
			return
		}
		if response.StatusCode != 200 {
			platform.Unavailable(w, fmt.Errorf("session validation returned %d", response.StatusCode))
			return
		}

		var session struct {
			UserID string `json:"userId"`
		}
		if err := json.NewDecoder(response.Body).Decode(&session); err != nil {
			platform.Unavailable(w, err)
			return
		}
		if !validID(session.UserID) {
			platform.Unavailable(w, errors.New("invalid session user ID"))
			return
		}
		next(w, r, session.UserID)
	}
}

func validID(id string) bool {
	parsed, err := uuid.Parse(id)
	return err == nil && parsed.String() == id
}

func resourceError(w http.ResponseWriter, err error) {
	if errors.Is(err, pgx.ErrNoRows) {
		platform.Error(w, 404, "not_found")
		return
	}
	platform.Unavailable(w, err)
}

func (s server) lockSong(ctx context.Context, tx pgx.Tx, id, owner string) (string, *string, error) {
	var status string
	var key *string
	err := tx.QueryRow(ctx, "SELECT status,file_key FROM song_library.songs WHERE id=$1 AND owner_id=$2 FOR UPDATE", id, owner).Scan(&status, &key)
	return status, key, err
}
