package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"golang.org/x/oauth2"

	"github.com/grachevelloper/frets/services/auth/internal/account"
)

type Provider struct {
	name, endpoint string
	config         oauth2.Config
}

func FromEnvironment(origin string) map[string]account.Provider {
	providers := make(map[string]account.Provider)
	for _, name := range []string{"google", "yandex"} {
		config, endpoint := providerConfig(origin, name)
		providers[name] = nil
		if config.ClientID != "" && config.ClientSecret != "" {
			providers[name] = Provider{name: name, endpoint: endpoint, config: config}
		}
	}
	return providers
}

func (p Provider) AuthorizationURL(state, verifier string) string {
	return p.config.AuthCodeURL(state, oauth2.S256ChallengeOption(verifier))
}

func providerConfig(origin, name string) (oauth2.Config, string) {
	config := oauth2.Config{
		ClientID:     os.Getenv(strings.ToUpper(name) + "_CLIENT_ID"),
		ClientSecret: os.Getenv(strings.ToUpper(name) + "_CLIENT_SECRET"),
		RedirectURL:  origin + "/api/v1/auth/callback/" + name,
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

func (p Provider) Identity(ctx context.Context, code, verifier string) (account.Identity, error) {
	access, err := p.config.Exchange(ctx, code, oauth2.VerifierOption(verifier))
	if err != nil {
		return account.Identity{}, err
	}
	req, err := http.NewRequestWithContext(ctx, "GET", p.endpoint, nil)
	if err != nil {
		return account.Identity{}, err
	}
	req.Header.Set("Authorization", "Bearer "+access.AccessToken)
	if p.name == "yandex" {
		req.Header.Set("Authorization", "OAuth "+access.AccessToken)
	}

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return account.Identity{}, err
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return account.Identity{}, fmt.Errorf("userinfo returned %d", response.StatusCode)
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
		return account.Identity{}, err
	}

	identity := account.Identity{ID: data.Sub, Name: data.Name, Email: data.Email}
	if p.name == "yandex" {
		if data.ClientID != p.config.ClientID {
			return account.Identity{}, errors.New("userinfo client mismatch")
		}
		identity = account.Identity{ID: data.ID, Name: data.DisplayName, Email: data.DefaultEmail}
		if identity.Name == "" {
			identity.Name = data.Login
		}
	}
	if identity.ID == "" {
		return account.Identity{}, errors.New("missing provider account ID")
	}
	if identity.Name == "" {
		identity.Name = p.name
	}
	return identity, nil
}
