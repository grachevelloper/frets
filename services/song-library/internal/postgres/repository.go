package postgres

import (
	"context"
	_ "embed"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/grachevelloper/frets/services/song-library/internal/library"
)

//go:embed schema.sql
var Schema string

type Repository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func notFound(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return library.ErrNotFound
	}
	return err
}

func lockSong(ctx context.Context, tx pgx.Tx, owner, id string) (string, *string, error) {
	var status string
	var key *string
	err := tx.QueryRow(ctx, "SELECT status,file_key FROM song_library.songs WHERE id=$1 AND owner_id=$2 FOR UPDATE", id, owner).Scan(&status, &key)
	return status, key, notFound(err)
}
