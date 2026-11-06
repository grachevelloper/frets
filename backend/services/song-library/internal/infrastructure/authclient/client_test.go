package authclient

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/grachevelloper/frets/backend/services/song-library/internal/domain/song"
)

func TestValidate(t *testing.T) {
	const userID = "04368ae3-0d49-48f4-aeb8-698eb9880480"
	for _, tt := range []struct {
		name         string
		code         int
		body         string
		want         string
		unauthorized bool
	}{
		{"valid", 200, `{"userId":"` + userID + `"}`, userID, false},
		{"revoked", 401, "", "", true},
		{"bad service secret", 403, "", "", false},
		{"bad response", 200, `{"userId":"invalid"}`, "", false},
		{"unavailable", 503, "", "", false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				cookie, err := r.Cookie("frets_session")
				if r.Method != "POST" || r.URL.Path != "/internal/sessions/validate" ||
					r.Header.Get("Authorization") != "Bearer secret" || err != nil || cookie.Value != "session" {
					t.Error("unexpected validation request")
				}
				w.WriteHeader(tt.code)
				w.Write([]byte(tt.body))
			}))
			defer upstream.Close()

			id, err := New(upstream.URL, "secret").Validate(t.Context(), "session")
			if id != tt.want || (err != nil) != (tt.want == "") || errors.Is(err, song.ErrUnauthorized) != tt.unauthorized {
				t.Fatalf("id = %q, err = %v", id, err)
			}
		})
	}
}
