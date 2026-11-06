package song

import (
	"context"
	"errors"
	"io"
	"reflect"
	"testing"

	domain "github.com/grachevelloper/frets/backend/services/song-library/internal/domain/song"
)

func TestResumePendingDeletion(t *testing.T) {
	repo := &repositoryStub{hidden: true}
	files := &storageStub{
		repo: repo,
		err:  errors.New("offline"),
	}
	u := New(repo, files)
	if err := u.ResumePendingDeletions(t.Context()); err == nil || repo.finished {
		t.Fatalf("failed deletion: finished=%v error=%v", repo.finished, err)
	}
	files.err = nil
	if err := u.ResumePendingDeletions(t.Context()); err != nil || !repo.finished {
		t.Fatalf("resumed deletion: finished=%v error=%v", repo.finished, err)
	}
	if err := u.ResumePendingDeletions(t.Context()); err != nil {
		t.Fatal(err)
	}
}

type canceledDeletionStorage struct {
	cancel context.CancelFunc
	calls  []string
}

func (s *canceledDeletionStorage) Open(context.Context, string) (io.ReadCloser, error) {
	return nil, errors.New("unexpected download")
}

func (s *canceledDeletionStorage) Remove(ctx context.Context, key string) error {
	s.calls = append(s.calls, key)
	s.cancel()
	return ctx.Err()
}

type pendingRepository struct {
	repositoryStub
	items       []domain.Deletion
	finishedIDs []string
}

func (r *pendingRepository) PendingDeletions(context.Context) ([]domain.Deletion, error) {
	return r.items, nil
}

func TestResumeStopsWhenApplicationIsCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	items := []domain.Deletion{
		{SongID: "first", FileKey: "first"},
		{SongID: "second", FileKey: "second"},
	}
	repo := &pendingRepository{items: items}
	files := &canceledDeletionStorage{cancel: cancel}
	err := New(repo, files).ResumePendingDeletions(ctx)
	if !errors.Is(err, context.Canceled) || !reflect.DeepEqual(files.calls, []string{"first"}) {
		t.Fatalf("err=%v removals=%v", err, files.calls)
	}
}

func (r *pendingRepository) FinishDeletion(_ context.Context, id string) error {
	r.finishedIDs = append(r.finishedIDs, id)
	return nil
}

type slowDeletionStorage struct{}

func (slowDeletionStorage) Open(context.Context, string) (io.ReadCloser, error) {
	return nil, errors.New("unexpected download")
}

func (slowDeletionStorage) Remove(ctx context.Context, key string) error {
	if key == "slow" {
		<-ctx.Done()
		return ctx.Err()
	}
	return ctx.Err()
}

func TestResumeContinuesAfterOneObjectTimesOut(t *testing.T) {
	items := []domain.Deletion{
		{SongID: "first", FileKey: "slow"},
		{SongID: "second", FileKey: "available"},
	}
	repo := &pendingRepository{items: items}
	err := New(repo, slowDeletionStorage{}).ResumePendingDeletions(t.Context())
	if !errors.Is(err, context.DeadlineExceeded) || !reflect.DeepEqual(repo.finishedIDs, []string{"second"}) {
		t.Fatalf("err=%v completed=%v", err, repo.finishedIDs)
	}
}
