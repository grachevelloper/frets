package upload

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/grachevelloper/frets/backend/services/song-library/internal/domain/song"
)

type uploadRepository struct {
	calls      *[]string
	attempt    string
	reserveErr error
	publishErr error
}

func (r uploadRepository) Reserve(_ context.Context, owner string, in song.UploadMetadata) (song.Song, string, error) {
	*r.calls = append(*r.calls, fmt.Sprintf("reserve %s %q %s %d", owner, in.Title, in.Key, in.Size))
	return song.Song{
		ID:    "song",
		Title: in.Title,
	}, r.attempt, r.reserveErr
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

func TestUpload(t *testing.T) {
	storageErr := errors.New("storage offline")
	const reserve = `reserve owner "  Song  " stable 4`
	tests := []struct {
		name       string
		attempt    string
		reserveErr error
		writeErr   error
		publishErr error
		wantErr    error
		invalid    bool
		created    bool
		want       []string
	}{
		{
			name:    "checksum mismatch has no side effects",
			invalid: true,
			wantErr: song.ErrChecksum,
		},
		{
			name:       "active attempt conflicts",
			reserveErr: song.ErrUploadInProgress,
			wantErr:    song.ErrUploadInProgress,
			want:       []string{reserve},
		},
		{
			name: "replay uses ready song",
			want: []string{reserve},
		},
		{
			name:    "publishes after writing",
			attempt: "attempt",
			created: true,
			want:    []string{reserve, "put attempt data", "publish song attempt"},
		},
		{
			name:     "failed write expires after cancellation",
			attempt:  "attempt",
			writeErr: storageErr,
			wantErr:  storageErr,
			want:     []string{reserve, "put attempt data", "expire attempt"},
		},
		{
			name:       "stale attempt cannot publish",
			attempt:    "attempt",
			publishErr: song.ErrUploadExpired,
			wantErr:    song.ErrUploadExpired,
			want:       []string{reserve, "put attempt data", "publish song attempt"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			var calls []string
			repo := uploadRepository{
				calls:      &calls,
				attempt:    tt.attempt,
				reserveErr: tt.reserveErr,
				publishErr: tt.publishErr,
			}
			files := fakeFiles{
				calls: &calls,
				err:   tt.writeErr,
			}
			if tt.writeErr != nil {
				files.cancel = cancel
			}
			data := []byte("data")
			in := Input{
				Title: "  Song  ",
				Key:   "stable",
				Hash:  fmt.Sprintf("%X", sha256.Sum256(data)),
				Data:  data,
			}
			if tt.invalid {
				in.Hash = "invalid"
			}

			result, err := New(repo, files).Upload(ctx, "owner", in)
			if !errors.Is(err, tt.wantErr) || errors.Is(err, context.Canceled) || result.Created != tt.created {
				t.Fatalf("created = %v, err = %v", result.Created, err)
			}
			if err == nil && result.Song.ID != "song" {
				t.Fatalf("song ID = %q", result.Song.ID)
			}
			if !reflect.DeepEqual(calls, tt.want) {
				t.Fatalf("calls = %v, want %v", calls, tt.want)
			}
		})
	}
}

func (r uploadRepository) ExpiredKeys(context.Context) ([]string, error) {
	return []string{"expired", "late"}, nil
}

func (f fakeFiles) Remove(_ context.Context, fileKey string) error {
	*f.calls = append(*f.calls, "remove "+fileKey)
	return f.err
}
