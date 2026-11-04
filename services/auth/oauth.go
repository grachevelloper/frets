package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"golang.org/x/oauth2"

	"github.com/grachevelloper/frets/services/internal/platform"
)

const browserCookie = "frets_login"

type identity struct {
	ID, Name string
	Email    *string
}

func (s server) provider(name string) (oauth2.Config, string) {
	config := oauth2.Config{
		ClientID:     os.Getenv(strings.ToUpper(name) + "_CLIENT_ID"),
		ClientSecret: os.Getenv(strings.ToUpper(name) + "_CLIENT_SECRET"),
		RedirectURL:  s.origin + "/api/v1/auth/callback/" + name,
	}
	switch name {
	case "google":
		config.Endpoint = oauth2.Endpoint{AuthURL: "https://accounts.google.com/o/oauth2/v2/auth", TokenURL: "https://oauth2.googleapis.com/token", AuthStyle: oauth2.AuthStyleInParams}
		config.Scopes = []string{"openid", "profile", "email"}
		return config, "https://openidconnect.googleapis.com/v1/userinfo"
	case "yandex":
		config.Endpoint = oauth2.Endpoint{AuthURL: "https://oauth.yandex.ru/authorize", TokenURL: "https://oauth.yandex.ru/token", AuthStyle: oauth2.AuthStyleInParams}
		config.Scopes = []string{"login:info", "login:email"}
		return config, "https://login.yandex.ru/info?format=json"
	default:
		return oauth2.Config{}, ""
	}
}

func returnPath(raw string) string {
	parsed, err := url.ParseRequestURI(raw)
	if err != nil || parsed.Host != "" || strings.Contains(parsed.Path, "\\") || strings.HasPrefix(parsed.Path, "//") {
		return "/"
	}
	if strings.HasPrefix(parsed.Path, "/songs/") {
		return parsed.EscapedPath()
	}
	return "/"
}

func (s server) login(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("provider")
	config, endpoint := s.provider(name)
	if endpoint == "" {
		platform.Error(w, 404, "not_found")
		return
	}
	if config.ClientID == "" || config.ClientSecret == "" {
		platform.Error(w, 503, "provider_unavailable")
		return
	}

	browser, state, verifier := token(), token(), oauth2.GenerateVerifier()
	ctx := r.Context()
	if _, err := s.db.Exec(ctx, "DELETE FROM auth.login_attempts WHERE expires_at <= now()"); err != nil {
		platform.Unavailable(w, err)
		return
	}
	_, err := s.db.Exec(ctx, `
 INSERT INTO auth.login_attempts VALUES($1,$2,$3,$4,$5,now()+interval '10 minutes')
 `, platform.Hash(state), name, platform.Hash(browser), verifier, returnPath(r.URL.Query().Get("next")))
	if err != nil {
		platform.Unavailable(w, err)
		return
	}

	s.cookie(w, browserCookie, browser, 600)
	http.Redirect(w, r, config.AuthCodeURL(state, oauth2.S256ChallengeOption(verifier)), 302)
}

func (s server) callback(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("provider")
	config, endpoint := s.provider(name)
	browser, err := r.Cookie(browserCookie)
	query := r.URL.Query()
	if endpoint == "" || err != nil || query.Get("state") == "" {
		platform.Error(w, 400, "login_invalid")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 12*time.Second)
	defer cancel()
	var verifier, next string
	err = s.db.QueryRow(ctx, `
 DELETE FROM auth.login_attempts
 WHERE state_hash=$1 AND provider=$2 AND browser_hash=$3 AND expires_at>now()
 RETURNING verifier,return_to
 `, platform.Hash(query.Get("state")), name, platform.Hash(browser.Value)).Scan(&verifier, &next)
	if errors.Is(err, pgx.ErrNoRows) {
		platform.Error(w, 400, "login_invalid")
		return
	}
	if err != nil {
		platform.Unavailable(w, err)
		return
	}

	s.cookie(w, browserCookie, "", -1)
	if query.Get("code") == "" {
		platform.Error(w, 400, "login_invalid")
		return
	}
	account, err := fetchIdentity(ctx, config, name, endpoint, query.Get("code"), verifier)
	if err != nil {
		platform.Error(w, 502, "login_failed")
		return
	}

	if err := s.createSession(ctx, w, name, account); err != nil {
		platform.Unavailable(w, err)
		return
	}
	http.Redirect(w, r, next, 302)
}

func fetchIdentity(ctx context.Context, config oauth2.Config, provider, endpoint, code, verifier string) (identity, error) {
	access, err := config.Exchange(ctx, code, oauth2.VerifierOption(verifier))
	if err != nil {
		return identity{}, err
	}
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return identity{}, err
	}
	req.Header.Set("Authorization", "Bearer "+access.AccessToken)
	if provider == "yandex" {
		req.Header.Set("Authorization", "OAuth "+access.AccessToken)
	}

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return identity{}, err
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return identity{}, fmt.Errorf("userinfo returned %d", response.StatusCode)
	}
	var data struct {
		Sub          string  `json:"sub"`
		ID           string  `json:"id"`
		Name         string  `json:"name"`
		Login        string  `json:"login"`
		DisplayName  string  `json:"display_name"`
		Email        *string `json:"email"`
		DefaultEmail *string `json:"default_email"`
		ClientID     string  `json:"client_id"`
	}
	if err := json.NewDecoder(response.Body).Decode(&data); err != nil {
		return identity{}, err
	}

	account := identity{ID: data.Sub, Name: data.Name, Email: data.Email}
	if provider == "yandex" {
		if data.ClientID != config.ClientID {
			return identity{}, errors.New("userinfo client mismatch")
		}
		account = identity{ID: data.ID, Name: data.DisplayName, Email: data.DefaultEmail}
		if account.Name == "" {
			account.Name = data.Login
		}
	}
	if account.ID == "" {
		return identity{}, errors.New("missing provider account ID")
	}
	if account.Name == "" {
		account.Name = provider
	}
	return account, nil
}
