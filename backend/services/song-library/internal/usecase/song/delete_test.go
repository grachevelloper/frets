package song

import (
	"context"
	"errors"
	"io"
	"testing"

	domain "github.com/grachevelloper/frets/backend/services/song-library/internal/domain/song"
)

type repositoryStub struct {
	hidden   bool
	finished bool
}

func (r *repositoryStub) BeginDeletion(context.Context, string, string) (string, error) {
	r.hidden = true
	if r.finished {
		return "", nil
	}
	return "object", nil
}

func (r *repositoryStub) FinishDeletion(context.Context, string) error {
	r.finished = true
	return nil
}

func (*repositoryStub) List(context.Context, string) ([]domain.Song, error) {
	return nil, domain.ErrNotFound
}

func (*repositoryStub) GetWithFragments(context.Context, string, string) (domain.WithFragments, error) {
	return domain.WithFragments{}, domain.ErrNotFound
}

func (*repositoryStub) File(context.Context, string, string) (string, int64, error) {
	return "", 0, domain.ErrNotFound
}

type storageStub struct {
	repo *repositoryStub
	err  error
}

func (s *storageStub) Open(context.Context, string) (io.ReadCloser, error) {
	return nil, errors.New("unexpected file access")
}

func (s *storageStub) Remove(context.Context, string) error {
	if !s.repo.hidden {
		return errors.New("removed object before hiding song")
	}
	return s.err
}

func TestDeleteRetriesSameOperation(t *testing.T) {
	repo := &repositoryStub{}
	files := &storageStub{
		repo: repo,
		err:  errors.New("offline"),
	}
	u := New(repo, files)
	pending, err := u.Delete(t.Context(), "owner", "song")
	if err != nil || !pending || !repo.hidden || repo.finished {
		t.Fatalf("pending=%v error=%v state=%+v", pending, err, repo)
	}
	files.err = nil
	pending, err = u.Delete(t.Context(), "owner", "song")
	if err != nil || pending || !repo.finished {
		t.Fatalf("pending=%v error=%v state=%+v", pending, err, repo)
	}
	if pending, err := u.Delete(t.Context(), "owner", "song"); err != nil || pending {
		t.Fatalf("completed replay: pending=%v error=%v", pending, err)
	}
}

func TestDownloadRejectsForeignSongBeforeStorage(t *testing.T) {
	u := New(&repositoryStub{}, nil)
	if _, _, err := u.Download(t.Context(), "other", "song"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal(err)
	}
}

func (r *repositoryStub) PendingDeletions(context.Context) ([]domain.Deletion, error) {
	if !r.hidden || r.finished {
		return nil, nil
	}
	return []domain.Deletion{{
		SongID:  "song",
		FileKey: "object",
	}}, nil
}
