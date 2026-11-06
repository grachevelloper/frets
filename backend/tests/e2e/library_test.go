package e2e

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

type api struct {
	t            *testing.T
	origin       string
	publicOrigin string
	client       *http.Client
}

type songResponse struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"createdAt"`
}

func newAPI(t *testing.T) *api {
	t.Helper()
	origin := os.Getenv("FRETS_E2E_ORIGIN")
	if origin == "" {
		t.Skip("FRETS_E2E_ORIGIN is unset")
	}
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	publicOrigin := os.Getenv("FRETS_E2E_PUBLIC_ORIGIN")
	if publicOrigin == "" {
		publicOrigin = origin
	}
	a := &api{
		t:            t,
		origin:       strings.TrimRight(origin, "/"),
		publicOrigin: strings.TrimRight(publicOrigin, "/"),
		client: &http.Client{
			Jar:     jar,
			Timeout: 10 * time.Second,
		},
	}
	a.request(http.MethodPost, "/auth/local", nil, nil, http.StatusNoContent)
	return a
}

func (a *api) request(method, path string, body []byte, headers map[string]string, want int) []byte {
	a.t.Helper()
	req, err := http.NewRequest(method, a.origin+"/api/v1"+path, bytes.NewReader(body))
	if err != nil {
		a.t.Fatal(err)
	}
	req.Header.Set("Origin", a.publicOrigin)
	for name, value := range headers {
		req.Header.Set(name, value)
	}
	response, err := a.client.Do(req)
	if err != nil {
		a.t.Fatal(err)
	}
	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	if err != nil {
		a.t.Fatal(err)
	}
	if response.StatusCode != want {
		a.t.Fatalf("%s %s: status=%d want=%d body=%s", method, path, response.StatusCode, want, data)
	}
	return data
}

func (a *api) upload(title, key string, data []byte, status int) songResponse {
	a.t.Helper()
	var body bytes.Buffer
	form := multipart.NewWriter(&body)
	sum := sha256.Sum256(data)
	if err := form.WriteField("title", title); err != nil {
		a.t.Fatal(err)
	}
	if err := form.WriteField("sha256", hex.EncodeToString(sum[:])); err != nil {
		a.t.Fatal(err)
	}
	file, err := form.CreateFormFile("file", "test.gp")
	if err != nil {
		a.t.Fatal(err)
	}
	if _, err := file.Write(data); err != nil {
		a.t.Fatal(err)
	}
	if err := form.Close(); err != nil {
		a.t.Fatal(err)
	}
	headers := map[string]string{
		"Content-Type":    form.FormDataContentType(),
		"Idempotency-Key": key,
	}
	response := a.request(http.MethodPost, "/songs", body.Bytes(), headers, status)
	var song songResponse
	if status == http.StatusCreated || status == http.StatusOK {
		if err := json.Unmarshal(response, &song); err != nil {
			a.t.Fatal(err)
		}
		id, err := uuid.Parse(song.ID)
		if err != nil || id.String() != song.ID {
			a.t.Fatalf("invalid song UUID %q", song.ID)
		}
	}
	return song
}

func TestLibrary(t *testing.T) {
	a := newAPI(t)
	const title = "  Учебная  "
	key := uuid.NewString()
	original := []byte("opaque score bytes")
	song := a.upload(title, key, original, http.StatusCreated)
	if song.ID == "" || song.Title != title {
		t.Fatalf("song=%+v", song)
	}
	path := "/songs/" + song.ID
	t.Cleanup(func() { a.request(http.MethodDelete, path, nil, nil, http.StatusNoContent) })
	replay := a.upload(title, key, original, http.StatusOK)
	if replay != song {
		t.Fatalf("replay=%+v original=%+v", replay, song)
	}
	a.upload("different title", key, original, http.StatusConflict)
	if got := a.request(http.MethodGet, path+"/file", nil, nil, http.StatusOK); !bytes.Equal(got, original) {
		t.Fatalf("downloaded=%q", got)
	}
	fragmentPath := path + "/fragments/" + uuid.NewString()
	bookmark := []byte(`{"title":"Припев","startBar":9,"endBar":12}`)
	headers := map[string]string{"Content-Type": "application/json"}
	first := a.request(http.MethodPut, fragmentPath, bookmark, headers, http.StatusCreated)
	second := a.request(http.MethodPut, fragmentPath, bookmark, headers, http.StatusOK)
	if !bytes.Equal(first, second) {
		t.Fatal("fragment changed on replay")
	}
	changed := []byte(`{"title":"Припев","startBar":9,"endBar":13}`)
	a.request(http.MethodPut, fragmentPath, changed, headers, http.StatusConflict)
	details := a.request(http.MethodGet, path, nil, nil, http.StatusOK)
	if !bytes.Contains(details, []byte("Припев")) {
		t.Fatalf("details=%s", details)
	}
	a.request(http.MethodDelete, path, nil, nil, http.StatusNoContent)
	a.request(http.MethodGet, path+"/file", nil, nil, http.StatusNotFound)
}
