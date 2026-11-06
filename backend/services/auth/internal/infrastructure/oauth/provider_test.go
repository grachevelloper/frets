package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"golang.org/x/oauth2"
)

func TestYandexAuthorizationURL(t *testing.T) {

	provider := NewYandex("http://localhost:6666", Credentials{
		ClientID:     "app",
		ClientSecret: "secret",
	})
	location, err := url.Parse(provider.AuthorizationURL("state", "verifier"))
	if err != nil {
		t.Fatal(err)
	}
	query := location.Query()
	if location.Host != "oauth.yandex.ru" || query.Get("redirect_uri") != "http://localhost:6666/api/v1/auth/callback/yandex" ||
		query.Get("state") != "state" || query.Get("code_challenge_method") != "S256" || query.Get("code_challenge") == "" {
		t.Fatalf("unexpected authorization URL: %s", location)
	}
}

func TestProviderIdentity(t *testing.T) {
	tests := []struct {
		name   string
		auth   string
		data   string
		wantID string
	}{
		{"google", "Bearer test-token", `{"sub":"google-id","id":"ignored","name":"G","email":"g@example.test"}`, "google-id"},
		{"yandex", "OAuth test-token", `{"id":"yandex-id","client_id":"app","display_name":"Y","default_email":"y@example.test"}`, "yandex-id"},
		{"google", "Bearer test-token", `{"sub":"google-id","name":"G"}`, ""},
		{"google", "Bearer test-token", `{"sub":"google-id","email":"invalid"}`, ""},
		{"yandex", "OAuth test-token", `{"id":"yandex-id","client_id":"app"}`, ""},
		{"yandex", "OAuth test-token", `{"id":"foreign-id","client_id":"another-app"}`, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
				if r.FormValue("code") != "code" || r.FormValue("code_verifier") != "verifier" {
					t.Error("token request lost code or PKCE verifier")
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]string{"access_token": "test-token"})
			})
			mux.HandleFunc("/info", func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Authorization") != tt.auth {
					t.Errorf("userinfo authorization = %q", r.Header.Get("Authorization"))
				}
				w.Write([]byte(tt.data))
			})
			server := httptest.NewServer(mux)
			defer server.Close()

			provider := Provider{
				name: tt.name, endpoint: server.URL + "/info",
				config: oauth2.Config{
					ClientID: "app",
					Endpoint: oauth2.Endpoint{
						TokenURL:  server.URL + "/token",
						AuthStyle: oauth2.AuthStyleInParams,
					},
				},
			}
			identity, err := provider.Identity(t.Context(), "code", "verifier")
			if (err != nil) != (tt.wantID == "") || identity.ID != tt.wantID {
				t.Fatalf("identity = %+v, err = %v", identity, err)
			}
		})
	}
}

func TestTokenExchangeErrorsDoNotExposeResponse(t *testing.T) {
	const secret = "private-token-and-account-data"
	responses := []struct {
		name   string
		body   string
		status int
	}{
		{
			name:   "OAuth error fields",
			body:   `{"error":"invalid_grant","error_description":"` + secret + `","error_uri":"https://example.test/` + secret + `"}`,
			status: http.StatusBadRequest,
		},
		{
			name:   "raw response body",
			body:   `{"access_token":"` + secret + `"}`,
			status: http.StatusBadRequest,
		},
		{
			name:   "malformed successful response",
			body:   `{"access_token":"token","expires_in":"` + secret + `"}`,
			status: http.StatusOK,
		},
	}
	for _, response := range responses {
		t.Run(response.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(response.status)
				fmt.Fprint(w, response.body)
			}))
			defer server.Close()
			provider := Provider{
				config: oauth2.Config{
					Endpoint: oauth2.Endpoint{
						TokenURL:  server.URL,
						AuthStyle: oauth2.AuthStyleInParams,
					},
				},
			}
			_, err := provider.Identity(t.Context(), "code", "verifier")
			if err == nil || strings.Contains(err.Error(), secret) {
				t.Fatalf("unsafe or missing exchange error: %v", err)
			}
			if response.status != http.StatusOK && !strings.Contains(err.Error(), fmt.Sprint(response.status)) {
				t.Fatalf("lost token endpoint status: %v", err)
			}
		})
	}
}

func TestTokenExchangePreservesCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	provider := NewGoogle("https://example.test", Credentials{ClientID: "app", ClientSecret: "secret"})
	_, err := provider.Identity(ctx, "code", "verifier")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("lost cancellation: %v", err)
	}
}
