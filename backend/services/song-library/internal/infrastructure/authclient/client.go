package authclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/grachevelloper/frets/backend/services/song-library/internal/domain/song"
)

type Client struct {
	validationURL string
	secret        string
	http          *http.Client
}

func New(baseURL, secret string) *Client {
	return &Client{
		validationURL: baseURL + "/internal/sessions/validate",
		secret:        secret,
		http: &http.Client{
			Timeout:       3 * time.Second,
			CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		},
	}
}

func (c *Client) Validate(ctx context.Context, token string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.validationURL, nil)
	if err != nil {
		return "", err
	}
	req.AddCookie(&http.Cookie{
		Name:  "frets_session",
		Value: token,
	})
	req.Header.Set("Authorization", "Bearer "+c.secret)

	response, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusUnauthorized {
		return "", song.ErrUnauthorized
	}
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("session validation returned %d", response.StatusCode)
	}

	var session struct {
		UserID string `json:"userId"`
	}
	if err := json.NewDecoder(response.Body).Decode(&session); err != nil {
		return "", err
	}
	id, err := uuid.Parse(session.UserID)
	if err != nil || id.String() != session.UserID {
		return "", errors.New("invalid session user ID")
	}
	return session.UserID, nil
}
