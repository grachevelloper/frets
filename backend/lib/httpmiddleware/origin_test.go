package httpmiddleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOrigin(t *testing.T) {
	const origin = "http://localhost:6666"
	cases := []struct {
		name   string
		method string
		origin string
		want   int
	}{
		{
			name:   "read",
			method: http.MethodGet,
			want:   http.StatusNoContent,
		},
		{
			name:   "same origin write",
			method: http.MethodPost,
			origin: origin,
			want:   http.StatusNoContent,
		},
		{
			name:   "foreign origin write",
			method: http.MethodPost,
			origin: "https://other.test",
			want:   http.StatusForbidden,
		},
		{
			name:   "missing origin write",
			method: http.MethodDelete,
			want:   http.StatusForbidden,
		},
	}
	handler := Origin(origin, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) }))
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/v1/resource", nil)
			req.Header.Set("Origin", tt.origin)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, req)
			if response.Code != tt.want {
				t.Fatalf("status=%d want=%d", response.Code, tt.want)
			}
		})
	}
}
