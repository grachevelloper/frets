package song

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestValidateUpload(t *testing.T) {
	data := []byte("score")
	valid := Upload{
		Title: "  Песня  ",
		Key:   "same-request",
		Data:  data,
		Hash:  fmt.Sprintf("%x", sha256.Sum256(data)),
	}
	tests := []struct {
		name   string
		change func(*Upload)
		want   error
	}{
		{
			name:   "preserves title",
			change: func(*Upload) {},
		},
		{
			name:   "missing key",
			change: func(in *Upload) { in.Key = "" },
			want:   ErrInvalidUpload,
		},
		{
			name:   "empty title",
			change: func(in *Upload) { in.Title = "  " },
			want:   ErrInvalidTitle,
		},
		{
			name:   "long unicode title",
			change: func(in *Upload) { in.Title = strings.Repeat("я", 201) },
			want:   ErrInvalidTitle,
		},
		{
			name:   "empty file",
			change: func(in *Upload) { in.Data = nil },
			want:   ErrInvalidUpload,
		},
		{
			name:   "checksum mismatch",
			change: func(in *Upload) { in.Hash = "wrong" },
			want:   ErrChecksum,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := valid
			tt.change(&in)
			metadata, err := ValidateUpload(in)
			if !errors.Is(err, tt.want) {
				t.Fatalf("error=%v want=%v", err, tt.want)
			}
			if err == nil && metadata.Title != valid.Title {
				t.Fatalf("title changed to %q", metadata.Title)
			}
		})
	}
}

func TestFragmentBounds(t *testing.T) {
	tests := []struct {
		start int32
		end   int32
		valid bool
	}{
		{
			start: 9,
			end:   12,
			valid: true,
		},
		{
			start: 1,
			end:   1,
			valid: true,
		},
		{
			start: 0,
			end:   1,
		},
		{
			start: 12,
			end:   9,
		},
	}
	for _, tt := range tests {
		input := FragmentInput{
			Title:    "Припев",
			StartBar: tt.start,
			EndBar:   tt.end,
		}
		if err := input.Validate(); (err == nil) != tt.valid {
			t.Errorf("range %d-%d: %v", tt.start, tt.end, err)
		}
	}
}
