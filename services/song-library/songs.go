package main

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/grachevelloper/frets/services/internal/platform"
)

type song struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"createdAt"`
}

type fragmentInput struct {
	Title    string `json:"title"`
	StartBar int32  `json:"startBar"`
	EndBar   int32  `json:"endBar"`
}

type fragment struct {
	ID     string `json:"id"`
	SongID string `json:"songId"`
	fragmentInput
}

func (s server) list(w http.ResponseWriter, r *http.Request, owner string) {
	rows, err := s.db.Query(r.Context(), `
 SELECT id,title,created_at FROM song_library.songs
 WHERE owner_id=$1 AND status='ready' ORDER BY created_at DESC,id
 `, owner)
	if err != nil {
		platform.Unavailable(w, err)
		return
	}
	defer rows.Close()

	items := make([]song, 0)
	for rows.Next() {
		var item song
		if err := rows.Scan(&item.ID, &item.Title, &item.CreatedAt); err != nil {
			platform.Unavailable(w, err)
			return
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		platform.Unavailable(w, err)
		return
	}
	platform.JSON(w, 200, items)
}

func (s server) detail(w http.ResponseWriter, r *http.Request, owner string) {
	id := r.PathValue("songId")
	if !validID(id) {
		platform.Error(w, 404, "not_found")
		return
	}

	ctx := r.Context()
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		platform.Unavailable(w, err)
		return
	}
	defer tx.Rollback(ctx)

	var item song
	err = tx.QueryRow(ctx, `
 SELECT id,title,created_at FROM song_library.songs WHERE id=$1 AND owner_id=$2 AND status='ready'
 `, id, owner).Scan(&item.ID, &item.Title, &item.CreatedAt)
	if err != nil {
		resourceError(w, err)
		return
	}

	rows, err := tx.Query(ctx, `
 SELECT id,title,start_bar,end_bar FROM song_library.fragments
 WHERE song_id=$1 ORDER BY start_bar,end_bar,id
 `, id)
	if err != nil {
		platform.Unavailable(w, err)
		return
	}
	defer rows.Close()
	fragments := make([]fragment, 0)
	for rows.Next() {
		f := fragment{SongID: id}
		if err := rows.Scan(&f.ID, &f.Title, &f.StartBar, &f.EndBar); err != nil {
			platform.Unavailable(w, err)
			return
		}
		fragments = append(fragments, f)
	}
	if err := rows.Err(); err != nil {
		platform.Unavailable(w, err)
		return
	}
	if err := tx.Commit(ctx); err != nil {
		platform.Unavailable(w, err)
		return
	}

	platform.JSON(w, 200, struct {
		song
		Fragments []fragment `json:"fragments"`
	}{item, fragments})
}

func (s server) download(w http.ResponseWriter, r *http.Request, owner string) {
	id := r.PathValue("songId")
	if !validID(id) {
		platform.Error(w, 404, "not_found")
		return
	}

	var key string
	var size int64
	ctx := r.Context()
	err := s.db.QueryRow(ctx, `
 SELECT file_key,size_bytes FROM song_library.songs WHERE id=$1 AND owner_id=$2 AND status='ready'
 `, id, owner).Scan(&key, &size)
	if err != nil {
		resourceError(w, err)
		return
	}

	file, err := s.files.open(ctx, key)
	if err != nil {
		platform.Unavailable(w, err)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", strconv.FormatInt(size, 10))
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if _, err := io.CopyN(w, file, size); err != nil {
		log.Print(err)
	}
}

func (s server) putFragment(w http.ResponseWriter, r *http.Request, owner string) {
	id, fragmentID := r.PathValue("songId"), r.PathValue("fragmentId")
	if !validID(id) {
		platform.Error(w, 404, "not_found")
		return
	}
	if !validID(fragmentID) {
		platform.Error(w, 400, "invalid_fragment")
		return
	}

	var input fragmentInput
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil || decoder.Decode(new(any)) != io.EOF ||
		!validTitle(input.Title) || input.StartBar < 1 || input.StartBar > input.EndBar {
		platform.Error(w, 400, "invalid_fragment")
		return
	}

	ctx := r.Context()
	tx, err := s.db.Begin(ctx)
	if err != nil {
		platform.Unavailable(w, err)
		return
	}
	defer tx.Rollback(ctx)

	status, _, err := s.lockSong(ctx, tx, id, owner)
	if err != nil {
		resourceError(w, err)
		return
	}
	if status != "ready" {
		platform.Error(w, 404, "not_found")
		return
	}

	result, err := tx.Exec(ctx, `
 INSERT INTO song_library.fragments(id,song_id,title,start_bar,end_bar)
 VALUES($1,$2,$3,$4,$5) ON CONFLICT(id) DO NOTHING
 `, fragmentID, id, input.Title, input.StartBar, input.EndBar)
	if err != nil {
		platform.Unavailable(w, err)
		return
	}

	response := fragment{fragmentID, id, input}
	code := 201
	if result.RowsAffected() == 0 {
		var previous fragmentInput
		err = tx.QueryRow(ctx, `
 SELECT title,start_bar,end_bar FROM song_library.fragments WHERE id=$1 AND song_id=$2
 `, fragmentID, id).Scan(&previous.Title, &previous.StartBar, &previous.EndBar)
		if errors.Is(err, pgx.ErrNoRows) || (err == nil && input != previous) {
			platform.Error(w, 409, "fragment_conflict")
			return
		}
		if err != nil {
			platform.Unavailable(w, err)
			return
		}
		code = 200
	}

	if err := tx.Commit(ctx); err != nil {
		platform.Unavailable(w, err)
		return
	}
	platform.JSON(w, code, response)
}
