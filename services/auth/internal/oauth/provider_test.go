package oauth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"golang.org/x/oauth2"
)

func TestYandexAuthorizationURL(t *testing.T) {
	t.Setenv("YANDEX_CLIENT_ID", "app")
	t.Setenv("YANDEX_CLIENT_SECRET", "secret")

	provider := FromEnvironment("http://localhost:6666")["yandex"]
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
		name, auth, data, wantID string
	}{
		{"google", "Bearer test-token", `{"sub":"google-id","id":"ignored","name":"G"}`, "google-id"},
		{"yandex", "OAuth test-token", `{"id":"yandex-id","client_id":"app","display_name":"Y"}`, "yandex-id"},
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
				config: oauth2.Config{ClientID: "app", Endpoint: oauth2.Endpoint{TokenURL: server.URL + "/token", AuthStyle: oauth2.AuthStyleInParams}},
			}
			identity, err := provider.Identity(t.Context(), "code", "verifier")
			if (err != nil) != (tt.wantID == "") || identity.ID != tt.wantID {
				t.Fatalf("identity = %+v, err = %v", identity, err)
			}
		})
	}
}
