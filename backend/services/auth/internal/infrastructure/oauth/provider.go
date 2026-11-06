package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"

	"github.com/grachevelloper/frets/backend/services/auth/internal/domain/auth"
)

type Provider struct {
	name     string
	endpoint string
	config   oauth2.Config
}

type Credentials struct {
	ClientID     string
	ClientSecret string
}

type tokenExchangeError struct {
	cause error
}

func (e tokenExchangeError) Error() string {
	var responseErr *oauth2.RetrieveError
	if errors.As(e.cause, &responseErr) && responseErr.Response != nil {
		return fmt.Sprintf("exchange OAuth token: HTTP %d", responseErr.Response.StatusCode)
	}
	return "exchange OAuth token failed"
}

func (e tokenExchangeError) Unwrap() error {
	return e.cause
}

func (p Provider) AuthorizationURL(state, verifier string) string {
	return p.config.AuthCodeURL(state, oauth2.S256ChallengeOption(verifier))
}

func (p Provider) Identity(ctx context.Context, code, verifier string) (auth.Identity, error) {
	access, err := p.config.Exchange(ctx, code, oauth2.VerifierOption(verifier))
	if err != nil {
		// Parsing errors can expose response fields even when the token endpoint returns HTTP 200.
		return auth.Identity{}, tokenExchangeError{cause: err}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.endpoint, nil)
	if err != nil {
		return auth.Identity{}, fmt.Errorf("create userinfo request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+access.AccessToken)
	if p.name == "yandex" {
		req.Header.Set("Authorization", "OAuth "+access.AccessToken)
	}

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return auth.Identity{}, fmt.Errorf("request userinfo: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return auth.Identity{}, fmt.Errorf("userinfo returned %d", response.StatusCode)
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
		return auth.Identity{}, fmt.Errorf("decode userinfo: %w", err)
	}

	identity := auth.Identity{
		ID:    data.Sub,
		Name:  data.Name,
		Email: data.Email,
	}
	if p.name == "yandex" {
		if data.ClientID != p.config.ClientID {
			return auth.Identity{}, errors.New("userinfo client mismatch")
		}
		identity = auth.Identity{
			ID:    data.ID,
			Name:  data.DisplayName,
			Email: data.DefaultEmail,
		}
		if identity.Name == "" {
			identity.Name = data.Login
		}
	}
	if identity.ID == "" {
		return auth.Identity{}, errors.New("missing provider account ID")
	}
	if identity.Name == "" {
		identity.Name = p.name
	}
	return identity, nil
}
