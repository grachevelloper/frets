package postgres

import (
	"context"

	"github.com/grachevelloper/frets/services/song-library/internal/library"
)

func (r *Repository) BeginDeletion(ctx context.Context, owner, id string) (string, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	status, key, err := lockSong(ctx, tx, owner, id)
	if err != nil {
		return "", err
	}
	switch status {
	case "deleted":
		return "", nil
	case "uploading":
		return "", library.ErrNotFound
	case "ready":
		if _, err := tx.Exec(ctx, "UPDATE song_library.songs SET status='deleting' WHERE id=$1", id); err != nil {
			return "", err
		}
	}
	return *key, tx.Commit(ctx)
}

func (r *Repository) FinishDeletion(ctx context.Context, id string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `
 UPDATE song_library.songs SET status='deleted',title=NULL,upload_key=NULL,sha256=NULL,size_bytes=NULL,file_key=NULL
 WHERE id=$1 AND status='deleting'
 `, id); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, "DELETE FROM song_library.fragments WHERE song_id=$1", id); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, "DELETE FROM song_library.upload_attempts WHERE song_id=$1 AND status='published'", id); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *Repository) ExpiredKeys(ctx context.Context) ([]string, error) {
	if _, err := r.db.Exec(ctx, `
 UPDATE song_library.upload_attempts SET status='expired'
 WHERE status='active' AND expires_at<=clock_timestamp()
 `); err != nil {
		return nil, err
	}

	// Revisit expired keys because an object store can finish a timed-out PUT after a successful DELETE.
	rows, err := r.db.Query(ctx, `
 WITH due AS (
 SELECT id FROM song_library.upload_attempts
 WHERE status='expired' AND sweep_after<=now() ORDER BY sweep_after LIMIT 32 FOR UPDATE SKIP LOCKED
 )
 UPDATE song_library.upload_attempts a SET sweep_after=now()+interval '1 minute'
 FROM due WHERE a.id=due.id RETURNING a.file_key
 `)
	if err != nil {
		return nil, err
	}
	var keys []string
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			rows.Close()
			return nil, err
		}
		keys = append(keys, key)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return keys, nil
}

func (r *Repository) PendingDeletions(ctx context.Context) ([]library.Deletion, error) {
	rows, err := r.db.Query(ctx, `
 SELECT id,file_key FROM song_library.songs WHERE status='deleting' ORDER BY created_at
 `)
	if err != nil {
		return nil, err
	}
	var pending []library.Deletion
	for rows.Next() {
		var item library.Deletion
		if err := rows.Scan(&item.SongID, &item.FileKey); err != nil {
			rows.Close()
			return nil, err
		}
		pending = append(pending, item)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return pending, nil
}
