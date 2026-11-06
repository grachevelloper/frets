package httpapi

import (
	"errors"
	"io"
	"net/http"

	"github.com/grachevelloper/frets/backend/lib/httpjson"
	"github.com/grachevelloper/frets/backend/services/song-library/internal/domain/song"
	"github.com/grachevelloper/frets/backend/services/song-library/internal/usecase/upload"
)

func (h Handler) Upload(w http.ResponseWriter, r *http.Request, owner string) {
	in, err := readUpload(w, r)
	if err != nil {
		failure(w, err)
		return
	}

	result, err := h.uploadUseCase.Upload(r.Context(), owner, in)
	if err != nil {
		failure(w, err)
		return
	}
	httpjson.Write(w, creationStatus(result.Created), toSongResponse(result.Song))
}

func readUpload(w http.ResponseWriter, r *http.Request) (upload.Input, error) {
	r.Body = http.MaxBytesReader(w, r.Body, song.MaxFileSizeBytes+(64<<10))
	if err := r.ParseMultipartForm(1 << 20); err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			return upload.Input{}, song.ErrFileTooLarge
		}
		return upload.Input{}, song.ErrInvalidUpload
	}
	defer r.MultipartForm.RemoveAll()

	file, _, err := r.FormFile("file")
	if err != nil {
		return upload.Input{}, song.ErrInvalidUpload
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, song.MaxFileSizeBytes+1))
	if err != nil {
		return upload.Input{}, song.ErrInvalidUpload
	}
	return upload.Input{
		Title: r.FormValue("title"),
		Hash:  r.FormValue("sha256"),
		Key:   r.Header.Get("Idempotency-Key"),
		Data:  data,
	}, nil
}
