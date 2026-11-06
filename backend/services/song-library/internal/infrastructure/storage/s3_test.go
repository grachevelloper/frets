package storage

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestS3ReadsPrivateObjectAndAcceptsMissingDelete(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" || r.URL.Path != "/scores/object" {
			t.Errorf("unexpected storage request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusForbidden)
			return
		}
		w.Header().Set("Last-Modified", "Mon, 02 Jan 2006 15:04:05 GMT")
		w.Header().Set("ETag", `"test-etag"`)
		switch r.Method {
		case http.MethodHead:
			w.Header().Set("Content-Length", "5")
			w.Header().Set("Last-Modified", "Mon, 02 Jan 2006 15:04:05 GMT")
			w.Header().Set("ETag", `"test-etag"`)
		case http.MethodGet:
			w.Header().Set("Content-Length", "5")
			w.Write([]byte("score"))
		case http.MethodDelete:
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`<Error><Code>NoSuchKey</Code><Message>Missing object</Message></Error>`))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	defer upstream.Close()
	storage, err := NewS3(S3Config{
		Endpoint:  strings.TrimPrefix(upstream.URL, "http://"),
		Bucket:    "scores",
		AccessKey: "access-key",
		SecretKey: "secret-key",
		Region:    "us-east-1",
		Insecure:  true,
	})
	if err != nil {
		t.Fatal(err)
	}
	object, err := storage.Open(t.Context(), "object")
	if err != nil {
		t.Fatal(err)
	}
	defer object.Close()
	data, err := io.ReadAll(object)
	if err != nil || string(data) != "score" {
		t.Fatalf("data=%q error=%v", data, err)
	}
	if err := storage.Remove(t.Context(), "object"); err != nil {
		t.Fatal(err)
	}
}
