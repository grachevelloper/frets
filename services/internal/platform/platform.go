package platform

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const SessionCookie = "frets_session"

var ErrUnauthorized = errors.New("unauthorized")

func Required(name string) string {
	value := os.Getenv(name)
	if value == "" {
		log.Fatalf("%s is required", name)
	}
	return value
}

func Database(schema, migration string) *pgxpool.Pool {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	db, err := pgxpool.New(ctx, Required("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	tx, err := db.Begin(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock(hashtext($1))", schema); err != nil {
		log.Fatal(err)
	}
	if _, err := tx.Exec(ctx, migration); err != nil {
		log.Fatal(err)
	}
	if err := tx.Commit(ctx); err != nil {
		log.Fatal(err)
	}
	return db
}

func JSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Print(err)
	}
}

func Error(w http.ResponseWriter, status int, code string) {
	JSON(w, status, map[string]string{"code": code})
}

func Unavailable(w http.ResponseWriter, err error) {
	log.Print(err)
	Error(w, http.StatusServiceUnavailable, "unavailable")
}

func Hash(value string) []byte {
	sum := sha256.Sum256([]byte(value))
	return sum[:]
}

func EqualSecret(a, b string) bool {
	return subtle.ConstantTimeCompare(Hash(a), Hash(b)) == 1
}

func Serve(mux *http.ServeMux) {
	origin := Required("PUBLIC_ORIGIN")
	server := &http.Server{
		Addr:              ":8080",
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       40 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       60 * time.Second,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "no-store")
			if (r.Method == "POST" || r.Method == "PUT" || r.Method == "DELETE") &&
				r.URL.Path != "/internal/sessions/validate" && r.Header.Get("Origin") != origin {
				Error(w, http.StatusForbidden, "origin_forbidden")
				return
			}
			if _, pattern := mux.Handler(r); pattern == "" {
				Error(w, http.StatusNotFound, "not_found")
				return
			}
			mux.ServeHTTP(w, r)
		}),
	}
	log.Fatal(fmt.Errorf("http server: %w", server.ListenAndServe()))
}
