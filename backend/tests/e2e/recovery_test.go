package e2e

import (
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func compose(t *testing.T, args ...string) []byte {
	t.Helper()
	project := os.Getenv("FRETS_E2E_COMPOSE_PROJECT")
	if project == "" {
		t.Skip("FRETS_E2E_COMPOSE_PROJECT is unset")
	}
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	command := []string{"compose", "--project-directory", root, "--env-file", filepath.Join(root, ".env.example"), "-p", project}
	output, err := exec.Command("docker", append(command, args...)...).CombinedOutput()
	if err != nil {
		t.Fatalf("compose %v: %v\n%s", args, err, output)
	}
	return output
}

func TestAuthOutageKeepsSession(t *testing.T) {
	if os.Getenv("FRETS_E2E_COMPOSE_PROJECT") == "" {
		t.Skip("FRETS_E2E_COMPOSE_PROJECT is unset")
	}
	a := newAPI(t)
	compose(t, "stop", "auth")
	t.Cleanup(func() { compose(t, "start", "auth") })
	a.request(http.MethodGet, "/songs", nil, nil, http.StatusServiceUnavailable)
	compose(t, "start", "auth")
	waitForProfile(t, a)
	a.request(http.MethodGet, "/songs", nil, nil, http.StatusOK)
}

func TestDeletionResumesAfterRestart(t *testing.T) {
	if os.Getenv("FRETS_E2E_COMPOSE_PROJECT") == "" {
		t.Skip("FRETS_E2E_COMPOSE_PROJECT is unset")
	}
	a := newAPI(t)
	song := a.upload("Recovery test", uuid.NewString(), []byte("recovery score"), http.StatusCreated)
	path := "/songs/" + song.ID
	compose(t, "exec", "-T", "-u", "0", "song-library", "chmod", "500", "/data/files")
	t.Cleanup(func() { compose(t, "exec", "-T", "-u", "0", "song-library", "chmod", "700", "/data/files") })
	a.request(http.MethodDelete, path, nil, nil, http.StatusAccepted)
	a.request(http.MethodGet, path, nil, nil, http.StatusNotFound)
	compose(t, "restart", "song-library")
	compose(t, "exec", "-T", "-u", "0", "song-library", "chmod", "700", "/data/files")
	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		// upload validates the UUID before it is included in psql's command string.
		status := compose(t, "exec", "-T", "postgres", "psql", "-U", "postgres", "-d", "frets", "-Atc",
			"SELECT status FROM song_library.songs WHERE id='"+song.ID+"'")
		if strings.TrimSpace(string(status)) == "deleted" {
			a.request(http.MethodDelete, path, nil, nil, http.StatusNoContent)
			return
		}
		time.Sleep(200 * time.Millisecond)
	}
	t.Fatal("deletion did not finish after restart")
}

func waitForProfile(t *testing.T, a *api) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		response, err := a.client.Get(a.origin + "/api/v1/me")
		if err == nil {
			response.Body.Close()
			if response.StatusCode == http.StatusOK {
				return
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatal("profile did not recover")
}
