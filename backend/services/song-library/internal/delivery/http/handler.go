package httpapi

import (
	"context"
	"io"
	"net/http"

	"github.com/google/uuid"

	"github.com/grachevelloper/frets/backend/lib/httpjson"
	"github.com/grachevelloper/frets/backend/services/song-library/internal/domain/song"
	"github.com/grachevelloper/frets/backend/services/song-library/internal/usecase/upload"
)

type SongUseCase interface {
	List(ctx context.Context, ownerID string) (items []song.Song, err error)
	GetWithFragments(ctx context.Context, ownerID string, songID string) (item song.WithFragments, err error)
	Download(ctx context.Context, ownerID string, songID string) (file io.ReadCloser, sizeBytes int64, err error)
	Delete(ctx context.Context, ownerID string, songID string) (pending bool, err error)
}

type UploadUseCase interface {
	Upload(ctx context.Context, ownerID string, input upload.Input) (result upload.Result, err error)
}

type FragmentUseCase interface {
	Create(ctx context.Context, ownerID string, fragment song.Fragment) (created bool, err error)
}

type Handler struct {
	songUseCase     SongUseCase
	uploadUseCase   UploadUseCase
	fragmentUseCase FragmentUseCase
}

func New(songUseCase SongUseCase, uploadUseCase UploadUseCase, fragmentUseCase FragmentUseCase) *Handler {
	return &Handler{
		songUseCase:     songUseCase,
		uploadUseCase:   uploadUseCase,
		fragmentUseCase: fragmentUseCase,
	}
}

func validID(id string) bool {
	parsed, err := uuid.Parse(id)
	return err == nil && parsed.String() == id
}

func songID(w http.ResponseWriter, r *http.Request) (string, bool) {
	id := r.PathValue("songId")
	if !validID(id) {
		httpjson.Error(w, http.StatusNotFound, "not_found")
		return "", false
	}
	return id, true
}
