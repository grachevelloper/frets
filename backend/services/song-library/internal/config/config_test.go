package config

import "testing"

func TestLoad(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://test")
	t.Setenv("PUBLIC_ORIGIN", "http://localhost:6666")
	t.Setenv("INTERNAL_SECRET", "secret")
	t.Setenv("AUTH_URL", "http://auth:8080")
	t.Setenv("FILE_DIR", t.TempDir())
	t.Setenv("S3_ENDPOINT", "")
	t.Setenv("S3_INSECURE", "false")
	cfg, err := Load()
	if err != nil || cfg.PublicOrigin != "http://localhost:6666" {
		t.Fatalf("config = %+v, error = %v", cfg, err)
	}
	invalidOrigins := []string{
		"https://example.test/unexpected/path",
		"https://example.test?",
	}
	for _, origin := range invalidOrigins {
		t.Setenv("PUBLIC_ORIGIN", origin)
		if _, err := Load(); err == nil {
			t.Fatalf("accepted invalid origin %q", origin)
		}
	}
	t.Setenv("PUBLIC_ORIGIN", "http://localhost:6666")
	invalidAuthURLs := []string{
		"http://auth:8080?debug=true",
		"http://auth:8080?",
		"http://auth:8080#fragment",
		"http://user:password@auth:8080",
	}
	for _, endpoint := range invalidAuthURLs {
		t.Setenv("AUTH_URL", endpoint)
		if _, err := Load(); err == nil {
			t.Fatalf("accepted invalid auth URL %q", endpoint)
		}
	}
	t.Setenv("AUTH_URL", "http://auth:8080")
	t.Setenv("INTERNAL_SECRET", "")
	if _, err := Load(); err == nil {
		t.Fatal("accepted missing internal secret")
	}
}
