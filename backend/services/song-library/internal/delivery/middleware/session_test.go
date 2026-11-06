package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/grachevelloper/frets/backend/services/song-library/internal/domain/song"
)

type validator struct {
	err error
}

func (v validator) Validate(context.Context, string) (string, error) {
	return "owner", v.err
}

func TestSessionFailures(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want int
		code string
	}{
		{
			name: "revoked",
			err:  song.ErrUnauthorized,
			want: http.StatusUnauthorized,
			code: "unauthorized",
		},
		{
			name: "auth outage",
			err:  errors.New("connection refused"),
			want: http.StatusServiceUnavailable,
			code: "unavailable",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			handler := Session(validator{err: tt.err}, func(http.ResponseWriter, *http.Request, string) { t.Fatal("called protected handler") })
			request := httptest.NewRequest(http.MethodGet, "/api/v1/songs", nil)
			request.AddCookie(&http.Cookie{
				Name:  "frets_session",
				Value: "token",
			})
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != tt.want || !strings.Contains(response.Body.String(), tt.code) {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
			}
		})
	}
}
