package library

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"reflect"
	"testing"
)

type uploadRepository struct {
	Repository
	calls                  *[]string
	attempt                string
	reserveErr, publishErr error
}

func (r uploadRepository) Reserve(_ context.Context, owner string, in UploadMetadata) (Song, string, error) {
	*r.calls = append(*r.calls, fmt.Sprintf("reserve %s %q %s %d", owner, in.Title, in.Key, in.Size))
	return Song{ID: "song", Title: in.Title}, r.attempt, r.reserveErr
}

func (r uploadRepository) Publish(_ context.Context, song, attempt string) error {
	*r.calls = append(*r.calls, "publish "+song+" "+attempt)
	return r.publishErr
}

func (r uploadRepository) ExpireAttempt(ctx context.Context, attempt string) error {
	*r.calls = append(*r.calls, "expire "+attempt)
	return ctx.Err()
}

type fakeFiles struct {
	Files
	calls  *[]string
	err    error
	cancel context.CancelFunc
}

func (f fakeFiles) Put(_ context.Context, key string, data []byte) error {
	*f.calls = append(*f.calls, "put "+key+" "+string(data))
	if f.cancel != nil {
		f.cancel()
	}
	return f.err
}

func (f fakeFiles) Remove(_ context.Context, key string) error {
	*f.calls = append(*f.calls, "remove "+key)
	return f.err
}

func TestUpload(t *testing.T) {
	storageErr := errors.New("storage offline")
	const reserve = `reserve owner "  Song  " stable 4`
	tests := []struct {
		name, attempt                             string
		reserveErr, writeErr, publishErr, wantErr error
		invalid, created                          bool
		want                                      []string
	}{
		{name: "checksum mismatch has no side effects", invalid: true, wantErr: ErrChecksum},
		{name: "active attempt conflicts", reserveErr: ErrUploadInProgress, wantErr: ErrUploadInProgress, want: []string{reserve}},
		{name: "replay uses ready song", want: []string{reserve}},
		{name: "publishes after writing", attempt: "attempt", created: true, want: []string{reserve, "put attempt data", "publish song attempt"}},
		{name: "failed write expires after cancellation", attempt: "attempt", writeErr: storageErr, wantErr: storageErr, want: []string{reserve, "put attempt data", "expire attempt"}},
		{name: "stale attempt cannot publish", attempt: "attempt", publishErr: ErrUploadExpired, wantErr: ErrUploadExpired, want: []string{reserve, "put attempt data", "publish song attempt"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			var calls []string
			repo := uploadRepository{calls: &calls, attempt: tt.attempt, reserveErr: tt.reserveErr, publishErr: tt.publishErr}
			files := fakeFiles{calls: &calls, err: tt.writeErr}
			if tt.writeErr != nil {
				files.cancel = cancel
			}
			data := []byte("data")
			in := Upload{Title: "  Song  ", Key: "stable", Hash: fmt.Sprintf("%X", sha256.Sum256(data)), Data: data}
			if tt.invalid {
				in.Hash = "invalid"
			}

			song, created, err := New(repo, files).Upload(ctx, "owner", in)
			if !errors.Is(err, tt.wantErr) || errors.Is(err, context.Canceled) || created != tt.created {
				t.Fatalf("created = %v, err = %v", created, err)
			}
			if err == nil && song.ID != "song" {
				t.Fatalf("song ID = %q", song.ID)
			}
			if !reflect.DeepEqual(calls, tt.want) {
				t.Fatalf("calls = %v, want %v", calls, tt.want)
			}
		})
	}
}

type deletionRepository struct {
	Repository
	calls *[]string
}

func (r deletionRepository) BeginDeletion(_ context.Context, owner, id string) (string, error) {
	*r.calls = append(*r.calls, "hide "+owner+" "+id)
	return "object", nil
}

func (r deletionRepository) PendingDeletions(context.Context) ([]Deletion, error) {
	return []Deletion{{SongID: "song", FileKey: "object"}}, nil
}

func (r deletionRepository) ExpiredKeys(context.Context) ([]string, error) {
	return []string{"old-attempt"}, nil
}

func (r deletionRepository) FinishDeletion(_ context.Context, id string) error {
	*r.calls = append(*r.calls, "finish "+id)
	return nil
}

func TestResumeDeletionAfterStorageFailure(t *testing.T) {
	var calls []string
	ctx := context.Background()
	repo := deletionRepository{calls: &calls}
	failing := New(repo, fakeFiles{calls: &calls, err: errors.New("storage offline")})

	pending, err := failing.Delete(ctx, "owner", "song")
	if err != nil || !pending {
		t.Fatalf("pending = %v, err = %v", pending, err)
	}

	restarted := New(repo, fakeFiles{calls: &calls})
	if err := restarted.ResumeDeletes(ctx); err != nil {
		t.Fatal(err)
	}
	if err := restarted.SweepUploads(ctx); err != nil {
		t.Fatal(err)
	}
	want := []string{"hide owner song", "remove object", "remove object", "finish song", "remove old-attempt"}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %v, want %v", calls, want)
	}
}

type inaccessibleRepository struct {
	Repository
}

func (inaccessibleRepository) File(context.Context, string, string) (string, int64, error) {
	return "", 0, ErrNotFound
}

func TestInaccessibleFileDoesNotReachStorage(t *testing.T) {
	var storage Files
	file, _, err := New(inaccessibleRepository{}, storage).Download(context.Background(), "other", "song")
	if file != nil || !errors.Is(err, ErrNotFound) {
		t.Fatalf("file = %v, err = %v", file, err)
	}
}

func TestFragmentBoundsRejectedBeforeRepository(t *testing.T) {
	service := New(nil, nil)
	for _, bounds := range [][2]int32{{0, 1}, {12, 9}} {
		f := Fragment{FragmentInput: FragmentInput{Title: "Припев", StartBar: bounds[0], EndBar: bounds[1]}}
		if _, err := service.PutFragment(context.Background(), "owner", f); !errors.Is(err, ErrInvalidFragment) {
			t.Errorf("bounds %v: err = %v", bounds, err)
		}
	}
}
