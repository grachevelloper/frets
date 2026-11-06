package httpapi

import (
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/grachevelloper/frets/backend/lib/httpjson"
)

func (h Handler) List(w http.ResponseWriter, r *http.Request, owner string) {
	items, err := h.songUseCase.List(r.Context(), owner)
	if err != nil {
		failure(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, toSongListResponse(items))
}

func (h Handler) GetSong(w http.ResponseWriter, r *http.Request, owner string) {
	id, ok := songID(w, r)
	if !ok {
		return
	}

	detail, err := h.songUseCase.GetWithFragments(r.Context(), owner, id)
	if err != nil {
		failure(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, toSongWithFragmentsResponse(detail))
}

func (h Handler) Download(w http.ResponseWriter, r *http.Request, owner string) {
	id, ok := songID(w, r)
	if !ok {
		return
	}

	file, size, err := h.songUseCase.Download(r.Context(), owner, id)
	if err != nil {
		failure(w, err)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", strconv.FormatInt(size, 10))
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if _, err := io.CopyN(w, file, size); err != nil {
		slog.Error("stream song file", "error", err)
	}
}

func (h Handler) Delete(w http.ResponseWriter, r *http.Request, owner string) {
	id, ok := songID(w, r)
	if !ok {
		return
	}

	pending, err := h.songUseCase.Delete(r.Context(), owner, id)
	if err != nil {
		failure(w, err)
		return
	}
	if pending {
		w.WriteHeader(http.StatusAccepted)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func creationStatus(created bool) int {
	if created {
		return http.StatusCreated
	}
	return http.StatusOK
}
