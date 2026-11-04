package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/grachevelloper/frets/services/internal/platform"
)

func (s server) delete(w http.ResponseWriter, r *http.Request, owner string) {
	id := r.PathValue("songId")
	if !validID(id) {
		platform.Error(w, 404, "not_found")
		return
	}

	ctx := r.Context()
	tx, err := s.db.Begin(ctx)
	if err != nil {
		platform.Unavailable(w, err)
		return
	}
	defer tx.Rollback(ctx)
	status, key, err := s.lockSong(ctx, tx, id, owner)
	if err != nil {
		resourceError(w, err)
		return
	}

	switch status {
	case "deleted":
		w.WriteHeader(204)
		return
	case "uploading":
		platform.Error(w, 404, "not_found")
		return
	case "ready":
		if _, err := tx.Exec(ctx, "UPDATE song_library.songs SET status='deleting' WHERE id=$1", id); err != nil {
			platform.Unavailable(w, err)
			return
		}
	}
	if err := tx.Commit(ctx); err != nil {
		platform.Unavailable(w, err)
		return
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := s.finishDelete(ctx, id, *key); err != nil {
		log.Printf("delete %s: %v", id, err)
		w.WriteHeader(202)
		return
	}
	w.WriteHeader(204)
}

func (s server) finishDelete(ctx context.Context, id, key string) error {
	if err := s.files.remove(ctx, key); err != nil {
		return err
	}

	tx, err := s.db.Begin(ctx)
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

func (s server) maintenance() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		if err := s.sweepUploads(ctx); err != nil {
			log.Printf("upload cleanup: %v", err)
		}
		cancel()

		ctx, cancel = context.WithTimeout(context.Background(), 20*time.Second)
		if err := s.resumeDeletes(ctx); err != nil {
			log.Printf("delete cleanup: %v", err)
		}
		cancel()
		<-ticker.C
	}
}

func (s server) sweepUploads(ctx context.Context) error {
	if _, err := s.db.Exec(ctx, `
 UPDATE song_library.upload_attempts SET status='expired'
 WHERE status='active' AND expires_at<=clock_timestamp()
 `); err != nil {
		return err
	}

	// Revisit expired keys because an object store can finish a timed-out PUT after a successful DELETE.
	rows, err := s.db.Query(ctx, `
 WITH due AS (
 SELECT id FROM song_library.upload_attempts
 WHERE status='expired' AND sweep_after<=now() ORDER BY sweep_after LIMIT 32 FOR UPDATE SKIP LOCKED
 )
 UPDATE song_library.upload_attempts a SET sweep_after=now()+interval '1 minute'
 FROM due WHERE a.id=due.id RETURNING a.file_key
 `)
	if err != nil {
		return err
	}
	var keys []string
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			rows.Close()
			return err
		}
		keys = append(keys, key)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	for _, key := range keys {
		itemCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		if err := s.files.remove(itemCtx, key); err != nil {
			log.Printf("sweep %s: %v", key, err)
		}
		cancel()
	}
	return nil
}

func (s server) resumeDeletes(ctx context.Context) error {
	rows, err := s.db.Query(ctx, `
 SELECT id,file_key FROM song_library.songs WHERE status='deleting' ORDER BY created_at
 `)
	if err != nil {
		return err
	}
	type deletion struct{ id, key string }
	var pending []deletion
	for rows.Next() {
		var item deletion
		if err := rows.Scan(&item.id, &item.key); err != nil {
			rows.Close()
			return err
		}
		pending = append(pending, item)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	for _, item := range pending {
		itemCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		if err := s.finishDelete(itemCtx, item.id, item.key); err != nil {
			log.Printf("delete %s: %v", item.id, err)
		}
		cancel()
	}
	return nil
}
