package httpapi

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/grachevelloper/frets/services/internal/platform"
	"github.com/grachevelloper/frets/services/song-library/internal/library"
)

func (h Handler) list(w http.ResponseWriter, r *http.Request, owner string) {
	items, err := h.library.List(r.Context(), owner)
	if err != nil {
		failure(w, err)
		return
	}
	platform.JSON(w, 200, items)
}

func (h Handler) detail(w http.ResponseWriter, r *http.Request, owner string) {
	id, ok := songID(w, r)
	if !ok {
		return
	}

	detail, err := h.library.Detail(r.Context(), owner, id)
	if err != nil {
		failure(w, err)
		return
	}
	platform.JSON(w, 200, detail)
}

func (h Handler) download(w http.ResponseWriter, r *http.Request, owner string) {
	id, ok := songID(w, r)
	if !ok {
		return
	}

	file, size, err := h.library.Download(r.Context(), owner, id)
	if err != nil {
		failure(w, err)
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

func (h Handler) putFragment(w http.ResponseWriter, r *http.Request, owner string) {
	id, ok := songID(w, r)
	if !ok {
		return
	}

	fragment := library.Fragment{ID: r.PathValue("fragmentId"), SongID: id}
	if !validID(fragment.ID) {
		failure(w, library.ErrInvalidFragment)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&fragment.FragmentInput); err != nil || decoder.Decode(new(any)) != io.EOF {
		failure(w, library.ErrInvalidFragment)
		return
	}

	created, err := h.library.PutFragment(r.Context(), owner, fragment)
	if err != nil {
		failure(w, err)
		return
	}
	platform.JSON(w, creationStatus(created), fragment)
}

func (h Handler) delete(w http.ResponseWriter, r *http.Request, owner string) {
	id, ok := songID(w, r)
	if !ok {
		return
	}

	pending, err := h.library.Delete(r.Context(), owner, id)
	if err != nil {
		failure(w, err)
		return
	}
	if pending {
		w.WriteHeader(202)
		return
	}
	w.WriteHeader(204)
}

func creationStatus(created bool) int {
	if created {
		return http.StatusCreated
	}
	return http.StatusOK
}
