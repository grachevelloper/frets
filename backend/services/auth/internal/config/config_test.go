package config

import "testing"

func TestLoad(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://test")
	t.Setenv("PUBLIC_ORIGIN", "http://localhost:6666")
	t.Setenv("INTERNAL_SECRET", "secret")
	t.Setenv("LOCAL_LOGIN", "true")
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
	t.Setenv("INTERNAL_SECRET", "")
	if _, err := Load(); err == nil {
		t.Fatal("accepted missing internal secret")
	}
}
