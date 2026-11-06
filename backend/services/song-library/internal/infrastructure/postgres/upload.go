package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/grachevelloper/frets/backend/services/song-library/internal/domain/song"
)

func (r *Repository) Reserve(ctx context.Context, owner string, in song.UploadMetadata) (song.Song, string, error) {
	var item song.Song
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return item, "", err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO song_library.songs(owner_id,title,status,upload_key,sha256,size_bytes)
		VALUES($1,$2,'uploading',$3,$4,$5) ON CONFLICT(owner_id,upload_key) DO NOTHING
	`, owner, in.Title, in.Key, in.Hash, in.Size)
	if err != nil {
		return item, "", err
	}

	var status song.Status
	var hash string
	var size int
	err = tx.QueryRow(ctx, `
		SELECT id,title,created_at,status,sha256,size_bytes FROM song_library.songs
		WHERE owner_id=$1 AND upload_key=$2 FOR UPDATE
	`, owner, in.Key).Scan(&item.ID, &item.Title, &item.CreatedAt, &status, &hash, &size)
	if errors.Is(err, pgx.ErrNoRows) {
		return item, "", song.ErrUploadConflict
	}
	if err != nil {
		return item, "", err
	}
	if item.Title != in.Title || hash != in.Hash || size != in.Size {
		return item, "", song.ErrUploadConflict
	}
	if status.Accessible() {
		return item, "", tx.Commit(ctx)
	}
	if !status.CanPublish() {
		return item, "", song.ErrUploadConflict
	}

	_, err = tx.Exec(ctx, `
		UPDATE song_library.upload_attempts SET status='expired'
		WHERE song_id=$1 AND status='active' AND expires_at<=clock_timestamp()
	`, item.ID)
	if err != nil {
		return item, "", err
	}

	attempt := uuid.NewString()
	result, err := tx.Exec(ctx, `
		INSERT INTO song_library.upload_attempts(id,song_id,file_key,status,expires_at)
		VALUES($1,$2,$3,'active',clock_timestamp()+interval '30 seconds')
		ON CONFLICT(song_id) WHERE status='active' DO NOTHING
	`, attempt, item.ID, attempt)
	if err != nil {
		return item, "", err
	}
	if result.RowsAffected() == 0 {
		return item, "", song.ErrUploadInProgress
	}
	return item, attempt, tx.Commit(ctx)
}

func (r *Repository) Publish(ctx context.Context, id, attempt string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var status song.Status
	if err := tx.QueryRow(ctx, "SELECT status FROM song_library.songs WHERE id=$1 FOR UPDATE", id).Scan(&status); err != nil {
		return err
	}
	if !status.CanPublish() {
		return song.ErrUploadExpired
	}

	result, err := tx.Exec(ctx, `
		UPDATE song_library.upload_attempts SET status='published'
		WHERE id=$1 AND song_id=$2 AND status='active' AND expires_at>clock_timestamp()
	`, attempt, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return song.ErrUploadExpired
	}

	if _, err := tx.Exec(ctx, "UPDATE song_library.songs SET status='ready',file_key=$2 WHERE id=$1", id, attempt); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *Repository) ExpireAttempt(ctx context.Context, attempt string) error {
	_, err := r.db.Exec(ctx, "UPDATE song_library.upload_attempts SET status='expired' WHERE id=$1 AND status='active'", attempt)
	return err
}
