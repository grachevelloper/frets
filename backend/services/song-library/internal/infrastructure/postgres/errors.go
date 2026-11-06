package postgres

import (
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/grachevelloper/frets/backend/services/song-library/internal/domain/song"
)

func mapRowError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return song.ErrNotFound
	}
	return err
}
