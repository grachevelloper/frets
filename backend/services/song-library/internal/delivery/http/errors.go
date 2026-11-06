package httpapi

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/grachevelloper/frets/backend/lib/httpjson"
	"github.com/grachevelloper/frets/backend/services/song-library/internal/domain/song"
)

func failure(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, song.ErrNotFound):
		httpjson.Error(w, http.StatusNotFound, "not_found")
	case errors.Is(err, song.ErrInvalidUpload):
		httpjson.Error(w, http.StatusBadRequest, "invalid_upload")
	case errors.Is(err, song.ErrInvalidTitle):
		httpjson.Error(w, http.StatusBadRequest, "invalid_title")
	case errors.Is(err, song.ErrChecksum):
		httpjson.Error(w, http.StatusBadRequest, "checksum_mismatch")
	case errors.Is(err, song.ErrInvalidFragment):
		httpjson.Error(w, http.StatusBadRequest, "invalid_fragment")
	case errors.Is(err, song.ErrFileTooLarge):
		httpjson.Error(w, http.StatusRequestEntityTooLarge, "file_too_large")
	case errors.Is(err, song.ErrUploadConflict):
		httpjson.Error(w, http.StatusConflict, "upload_conflict")
	case errors.Is(err, song.ErrUploadInProgress):
		httpjson.Error(w, http.StatusConflict, "upload_in_progress")
	case errors.Is(err, song.ErrUploadExpired):
		httpjson.Error(w, http.StatusConflict, "upload_expired")
	case errors.Is(err, song.ErrFragmentConflict):
		httpjson.Error(w, http.StatusConflict, "fragment_conflict")
	default:
		slog.Error("library request failed", "error", err)
		httpjson.Error(w, http.StatusServiceUnavailable, "unavailable")
	}
}

func NotFound(w http.ResponseWriter, r *http.Request) {
	httpjson.Error(w, http.StatusNotFound, "not_found")
}
