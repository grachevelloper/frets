package httpapi

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/grachevelloper/frets/backend/services/song-library/internal/domain/song"
)

func TestWrappedErrorsKeepStablePublicCodes(t *testing.T) {
	cases := []struct {
		err    error
		status int
		code   string
	}{
		{
			err:    song.ErrUploadConflict,
			status: http.StatusConflict,
			code:   "upload_conflict",
		},
		{
			err:    song.ErrInvalidFragment,
			status: http.StatusBadRequest,
			code:   "invalid_fragment",
		},
		{
			err:    song.ErrNotFound,
			status: http.StatusNotFound,
			code:   "not_found",
		},
	}
	for _, tt := range cases {
		response := httptest.NewRecorder()
		failure(response, fmt.Errorf("private database details: %w", tt.err))
		body := response.Body.String()
		if response.Code != tt.status || !strings.Contains(body, tt.code) || strings.Contains(body, "private") {
			t.Fatalf("status=%d body=%s", response.Code, body)
		}
	}
}
