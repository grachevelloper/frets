package httpmiddleware

import (
	"net/http"

	"github.com/grachevelloper/frets/backend/lib/httpjson"
)

func Origin(origin string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		modifying := r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch || r.Method == http.MethodDelete
		if modifying && r.Header.Get("Origin") != origin {
			httpjson.Error(w, http.StatusForbidden, "origin_forbidden")
			return
		}
		next.ServeHTTP(w, r)
	})
}
