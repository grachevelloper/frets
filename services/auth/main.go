package main

import (
	"context"
	"crypto/rand"
	_ "embed"
	"encoding/base64"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/grachevelloper/frets/services/internal/platform"
)

//go:embed schema.sql
var schema string

type server struct {
	db                     *pgxpool.Pool
	origin, internalSecret string
	secure, localLogin     bool
}

func main() {
	origin := platform.Required("PUBLIC_ORIGIN")
	s := server{
		db:             platform.Database("auth", schema),
		origin:         origin,
		internalSecret: platform.Required("INTERNAL_SECRET"),
		secure:         strings.HasPrefix(origin, "https://"),
		localLogin:     origin == "http://localhost:6666" && os.Getenv("LOCAL_LOGIN") == "true",
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/auth/login/{provider}", s.login)
	mux.HandleFunc("GET /api/v1/auth/callback/{provider}", s.callback)
	mux.HandleFunc("GET /api/v1/me", s.me)
	mux.HandleFunc("POST /api/v1/auth/logout", s.logout)
	mux.HandleFunc("POST /internal/sessions/validate", s.validate)
	if s.localLogin {
		mux.HandleFunc("POST /api/v1/auth/local", s.loginLocal)
	}
	platform.Serve(mux)
}

func token() string {
	var data [32]byte
	if _, err := rand.Read(data[:]); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(data[:])
}

func (s server) cookie(w http.ResponseWriter, name, value string, maxAge int) {
	http.SetCookie(w, &http.Cookie{
		Name: name, Value: value, Path: "/", MaxAge: maxAge,
		HttpOnly: true, Secure: s.secure, SameSite: http.SameSiteLaxMode,
	})
}

func sessionToken(r *http.Request) string {
	cookie, err := r.Cookie(platform.SessionCookie)
	if err != nil {
		return ""
	}
	return cookie.Value
}

type profile struct {
	ID        string    `json:"id"`
	Nickname  string    `json:"nickname"`
	Email     *string   `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
}

func (s server) session(ctx context.Context, token string) (profile, error) {
	var p profile
	if token == "" {
		return p, pgx.ErrNoRows
	}
	err := s.db.QueryRow(ctx, `
 SELECT u.id, u.nickname, u.email, u.created_at
 FROM auth.sessions s JOIN auth.users u ON u.id = s.user_id
 WHERE s.token_hash = $1 AND s.revoked_at IS NULL AND s.expires_at > now()
 `, platform.Hash(token)).Scan(&p.ID, &p.Nickname, &p.Email, &p.CreatedAt)
	return p, err
}

func sessionError(w http.ResponseWriter, err error) {
	if errors.Is(err, pgx.ErrNoRows) {
		platform.Error(w, 401, "unauthorized")
		return
	}
	platform.Unavailable(w, err)
}

func (s server) me(w http.ResponseWriter, r *http.Request) {
	p, err := s.session(r.Context(), sessionToken(r))
	if err != nil {
		sessionError(w, err)
		return
	}
	platform.JSON(w, 200, p)
}

func (s server) validate(w http.ResponseWriter, r *http.Request) {
	if !platform.EqualSecret(r.Header.Get("Authorization"), "Bearer "+s.internalSecret) {
		platform.Error(w, 403, "forbidden")
		return
	}
	p, err := s.session(r.Context(), sessionToken(r))
	if err != nil {
		sessionError(w, err)
		return
	}
	platform.JSON(w, 200, map[string]string{"userId": p.ID})
}

func (s server) logout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	result, err := s.db.Exec(ctx, `
 UPDATE auth.sessions SET revoked_at = now()
 WHERE token_hash = $1 AND revoked_at IS NULL AND expires_at > now()
 `, platform.Hash(sessionToken(r)))
	if err != nil {
		platform.Unavailable(w, err)
		return
	}
	s.cookie(w, platform.SessionCookie, "", -1)
	if result.RowsAffected() == 0 {
		platform.Error(w, 401, "unauthorized")
		return
	}
	w.WriteHeader(204)
}

func (s server) createSession(ctx context.Context, w http.ResponseWriter, provider string, account identity) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Serialize first sign-in for this identity so concurrent callbacks create only one profile.
	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock(hashtextextended($1, 1))", provider+":"+account.ID); err != nil {
		return err
	}

	var userID string
	err = tx.QueryRow(ctx, "SELECT user_id FROM auth.identities WHERE provider=$1 AND provider_account_id=$2", provider, account.ID).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		err = tx.QueryRow(ctx, "INSERT INTO auth.users(nickname,email) VALUES($1,$2) RETURNING id", account.Name, account.Email).Scan(&userID)
		if err == nil {
			_, err = tx.Exec(ctx, "INSERT INTO auth.identities VALUES($1,$2,$3)", provider, account.ID, userID)
		}
	}
	if err != nil {
		return err
	}

	key := token()
	if _, err := tx.Exec(ctx, "INSERT INTO auth.sessions(token_hash,user_id,expires_at) VALUES($1,$2,now()+interval '30 days')", platform.Hash(key), userID); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	s.cookie(w, platform.SessionCookie, key, 30*24*60*60)
	return nil
}

func (s server) loginLocal(w http.ResponseWriter, r *http.Request) {
	account := identity{ID: "local", Name: "Local guitarist"}
	if err := s.createSession(r.Context(), w, "local", account); err != nil {
		platform.Unavailable(w, err)
		return
	}
	w.WriteHeader(204)
}
