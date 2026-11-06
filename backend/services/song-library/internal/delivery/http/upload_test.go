package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/grachevelloper/frets/backend/services/song-library/internal/domain/song"
	"github.com/grachevelloper/frets/backend/services/song-library/internal/usecase/upload"
)

type uploadStub struct {
	input   upload.Input
	created bool
}

func (s *uploadStub) Upload(_ context.Context, owner string, in upload.Input) (upload.Result, error) {
	s.input = in
	return upload.Result{
		Song: song.Song{
			ID:    "song-id",
			Title: in.Title,
		},
		Created: s.created,
	}, nil
}

func TestUploadMultipartAndResponse(t *testing.T) {
	var body bytes.Buffer
	form := multipart.NewWriter(&body)
	if err := form.WriteField("title", "  Песня  "); err != nil {
		t.Fatal(err)
	}
	if err := form.WriteField("sha256", "digest"); err != nil {
		t.Fatal(err)
	}
	file, err := form.CreateFormFile("file", "song.gp")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.Write([]byte("score")); err != nil {
		t.Fatal(err)
	}
	if err := form.Close(); err != nil {
		t.Fatal(err)
	}

	scenarios := []struct {
		name    string
		created bool
		status  int
	}{
		{
			name:    "new upload",
			created: true,
			status:  http.StatusCreated,
		},
		{
			name:   "successful replay",
			status: http.StatusOK,
		},
	}
	for _, tt := range scenarios {
		t.Run(tt.name, func(t *testing.T) {
			uploads := &uploadStub{created: tt.created}
			h := New(nil, uploads, nil)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/songs", bytes.NewReader(body.Bytes()))
			req.Header.Set("Content-Type", form.FormDataContentType())
			req.Header.Set("Idempotency-Key", "same-request")
			response := httptest.NewRecorder()
			h.Upload(response, req, "owner")
			var result map[string]json.RawMessage
			if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			if response.Code != tt.status || uploads.input.Key != "same-request" || string(uploads.input.Data) != "score" || uploads.input.Title != "  Песня  " || result["createdAt"] == nil {
				t.Fatalf("status=%d body=%s input=%+v", response.Code, response.Body.String(), uploads.input)
			}
		})
	}
}
