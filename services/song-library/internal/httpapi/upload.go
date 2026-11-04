package httpapi

import (
	"errors"
	"io"
	"net/http"

	"github.com/grachevelloper/frets/services/internal/platform"
	"github.com/grachevelloper/frets/services/song-library/internal/library"
)

func (h Handler) upload(w http.ResponseWriter, r *http.Request, owner string) {
	in, err := readUpload(w, r)
	if err != nil {
		failure(w, err)
		return
	}

	song, created, err := h.library.Upload(r.Context(), owner, in)
	if err != nil {
		failure(w, err)
		return
	}
	platform.JSON(w, creationStatus(created), song)
}

func readUpload(w http.ResponseWriter, r *http.Request) (library.Upload, error) {
	r.Body = http.MaxBytesReader(w, r.Body, library.MaxFileSize+(64<<10))
	if err := r.ParseMultipartForm(1 << 20); err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			return library.Upload{}, library.ErrFileTooLarge
		}
		return library.Upload{}, library.ErrInvalidUpload
	}
	defer r.MultipartForm.RemoveAll()

	file, _, err := r.FormFile("file")
	if err != nil {
		return library.Upload{}, library.ErrInvalidUpload
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, library.MaxFileSize+1))
	if err != nil {
		return library.Upload{}, library.ErrInvalidUpload
	}
	return library.Upload{
		Title: r.FormValue("title"),
		Hash:  r.FormValue("sha256"),
		Key:   r.Header.Get("Idempotency-Key"),
		Data:  data,
	}, nil
}
