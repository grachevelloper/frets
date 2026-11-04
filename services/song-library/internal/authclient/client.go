package authclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/grachevelloper/frets/services/internal/platform"
)

type Client struct {
	url, secret string
	http        *http.Client
}

func New(url, secret string) *Client {
	return &Client{
		url:    url + "/internal/sessions/validate",
		secret: secret,
		http: &http.Client{
			Timeout:       3 * time.Second,
			CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		},
	}
}

func (c *Client) Validate(ctx context.Context, token string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", c.url, nil)
	if err != nil {
		return "", err
	}
	req.AddCookie(&http.Cookie{Name: platform.SessionCookie, Value: token})
	req.Header.Set("Authorization", "Bearer "+c.secret)

	response, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	switch response.StatusCode {
	case 401:
		return "", platform.ErrUnauthorized
	case 200:
	default:
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
