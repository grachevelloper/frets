package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/grachevelloper/frets/services/internal/platform"
)

const maxFile = 16 << 20

var (
	errConflict   = errors.New("upload_conflict")
	errInProgress = errors.New("upload_in_progress")
	errExpired    = errors.New("upload_expired")
)

type uploadInput struct {
	title, key, hash string
	data             []byte
}

func validTitle(title string) bool {
	return utf8.ValidString(title) && !strings.ContainsRune(title, 0) && strings.TrimSpace(title) != "" && utf8.RuneCountInString(title) <= 200
}

func readUpload(w http.ResponseWriter, r *http.Request) (uploadInput, bool) {
	var in uploadInput
	in.key = r.Header.Get("Idempotency-Key")
	if in.key == "" || len(in.key) > 128 || !utf8.ValidString(in.key) || strings.ContainsRune(in.key, 0) {
		platform.Error(w, 400, "invalid_upload")
		return in, false
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxFile+(64<<10))
	if err := r.ParseMultipartForm(1 << 20); err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			platform.Error(w, 413, "file_too_large")
		} else {
			platform.Error(w, 400, "invalid_upload")
		}
		return in, false
	}
	defer r.MultipartForm.RemoveAll()

	in.title = r.FormValue("title")
	in.hash = strings.ToLower(r.FormValue("sha256"))
	file, _, err := r.FormFile("file")
	if err != nil {
		platform.Error(w, 400, "invalid_upload")
		return in, false
	}
	defer file.Close()
	if !validTitle(in.title) {
		platform.Error(w, 400, "invalid_title")
		return in, false
	}

	in.data, err = io.ReadAll(io.LimitReader(file, maxFile+1))
	if err != nil || len(in.data) == 0 {
		platform.Error(w, 400, "invalid_upload")
		return in, false
	}
	if len(in.data) > maxFile {
		platform.Error(w, 413, "file_too_large")
		return in, false
	}

	actual := sha256.Sum256(in.data)
	if hex.EncodeToString(actual[:]) != in.hash {
		platform.Error(w, 400, "checksum_mismatch")
		return in, false
	}
	return in, true
}

func (s server) upload(w http.ResponseWriter, r *http.Request, owner string) {
	in, ok := readUpload(w, r)
	if !ok {
		return
	}
	ctx := r.Context()
	item, attempt, err := s.reserve(ctx, owner, in)
	if err != nil {
		uploadError(w, err)
		return
	}
	if attempt == "" {
		platform.JSON(w, 200, item)
		return
	}

	writeCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	err = s.files.put(writeCtx, attempt, in.data)
	cancel()
	if err != nil {
		releaseCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 3*time.Second)
		defer cancel()
		if _, releaseErr := s.db.Exec(releaseCtx, "UPDATE song_library.upload_attempts SET status='expired' WHERE id=$1 AND status='active'", attempt); releaseErr != nil {
			platform.Unavailable(w, errors.Join(err, releaseErr))
			return
		}
		platform.Unavailable(w, err)
		return
	}
	if err := s.publish(ctx, item.ID, attempt); err != nil {
		uploadError(w, err)
		return
	}
	platform.JSON(w, 201, item)
}

func uploadError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, errConflict), errors.Is(err, errInProgress), errors.Is(err, errExpired):
		platform.Error(w, 409, err.Error())
	default:
		platform.Unavailable(w, err)
	}
}

func (s server) reserve(ctx context.Context, owner string, in uploadInput) (song, string, error) {
	var item song
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return item, "", err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
 INSERT INTO song_library.songs(owner_id,title,status,upload_key,sha256,size_bytes)
 VALUES($1,$2,'uploading',$3,$4,$5) ON CONFLICT(owner_id,upload_key) DO NOTHING
 `, owner, in.title, in.key, in.hash, len(in.data))
	if err != nil {
		return item, "", err
	}

	var status, hash string
	var size int
	err = tx.QueryRow(ctx, `
 SELECT id,title,created_at,status,sha256,size_bytes FROM song_library.songs
 WHERE owner_id=$1 AND upload_key=$2 FOR UPDATE
 `, owner, in.key).Scan(&item.ID, &item.Title, &item.CreatedAt, &status, &hash, &size)
	if errors.Is(err, pgx.ErrNoRows) {
		return item, "", errConflict
	}
	if err != nil {
		return item, "", err
	}
	if item.Title != in.title || hash != in.hash || size != len(in.data) {
		return item, "", errConflict
	}
	if status == "ready" {
		return item, "", tx.Commit(ctx)
	}
	if status != "uploading" {
		return item, "", errConflict
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
		return item, "", errInProgress
	}
	return item, attempt, tx.Commit(ctx)
}

func (s server) publish(ctx context.Context, id, attempt string) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var status string
	if err := tx.QueryRow(ctx, "SELECT status FROM song_library.songs WHERE id=$1 FOR UPDATE", id).Scan(&status); err != nil {
		return err
	}
	if status != "uploading" {
		return errExpired
	}

	result, err := tx.Exec(ctx, `
 UPDATE song_library.upload_attempts SET status='published'
 WHERE id=$1 AND song_id=$2 AND status='active' AND expires_at>clock_timestamp()
 `, attempt, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return errExpired
	}

	if _, err := tx.Exec(ctx, "UPDATE song_library.songs SET status='ready',file_key=$2 WHERE id=$1", id, attempt); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
