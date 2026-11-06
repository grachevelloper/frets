package delivery

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	httpapi "github.com/grachevelloper/frets/backend/services/auth/internal/delivery/http"
	"github.com/grachevelloper/frets/backend/services/auth/internal/domain/auth"
)

type sessionStub struct{}

func (sessionStub) Create(context.Context, string, auth.Identity) (string, error) {
	return "token", nil
}
func (sessionStub) Profile(context.Context, string) (auth.Profile, error) {
	return auth.Profile{ID: "owner"}, nil
}
func (sessionStub) Revoke(context.Context, string) error { return nil }

func TestOriginAndInternalRoutes(t *testing.T) {
	const origin = "http://localhost:6666"
	handler := httpapi.New(sessionStub{}, nil, httpapi.CookieConfig{InternalSecret: "service-secret"})
	router := NewRouter(origin, handler, false)
	cases := []struct {
		name       string
		path       string
		origin     string
		credential string
		status     int
		code       string
	}{
		{
			name:       "internal validation",
			path:       "/internal/sessions/validate",
			credential: "Bearer service-secret",
			status:     http.StatusOK,
		},
		{
			name:   "internal credential required",
			path:   "/internal/sessions/validate",
			status: http.StatusForbidden,
			code:   "forbidden",
		},
		{
			name:   "public write requires origin",
			path:   "/api/v1/auth/logout",
			status: http.StatusForbidden,
			code:   "origin_forbidden",
		},
		{
			name:   "public write rejects foreign origin",
			path:   "/api/v1/auth/logout",
			origin: "https://other.test",
			status: http.StatusForbidden,
			code:   "origin_forbidden",
		},
		{
			name:   "public write accepts configured origin",
			path:   "/api/v1/auth/logout",
			origin: origin,
			status: http.StatusNoContent,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, tt.path, nil)
			req.Header.Set("Origin", tt.origin)
			req.Header.Set("Authorization", tt.credential)
			req.AddCookie(&http.Cookie{
				Name:  "frets_session",
				Value: "token",
			})
			response := httptest.NewRecorder()
			router.ServeHTTP(response, req)
			if response.Code != tt.status || (tt.code != "" && !strings.Contains(response.Body.String(), tt.code)) {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
			}
		})
	}
}
