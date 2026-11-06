package httpapi

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/grachevelloper/frets/backend/lib/httpjson"
	"github.com/grachevelloper/frets/backend/services/song-library/internal/domain/song"
)

func (h Handler) CreateFragment(w http.ResponseWriter, r *http.Request, owner string) {
	id, ok := songID(w, r)
	if !ok {
		return
	}

	fragment := song.Fragment{
		ID:     r.PathValue("fragmentId"),
		SongID: id,
	}
	if !validID(fragment.ID) {
		failure(w, song.ErrInvalidFragment)
		return
	}

	var input fragmentRequest
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil || decoder.Decode(new(any)) != io.EOF {
		failure(w, song.ErrInvalidFragment)
		return
	}

	fragment.FragmentInput = song.FragmentInput{
		Title:    input.Title,
		StartBar: input.StartBar,
		EndBar:   input.EndBar,
	}

	created, err := h.fragmentUseCase.Create(r.Context(), owner, fragment)
	if err != nil {
		failure(w, err)
		return
	}
	httpjson.Write(w, creationStatus(created), toFragmentResponse(fragment))
}
