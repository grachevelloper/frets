package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/grachevelloper/frets/backend/services/auth/internal/domain/auth"
)

type sessionsStub struct {
	revoked bool
}

func (*sessionsStub) Create(context.Context, string, auth.Identity) (string, error) {
	return "token", nil
}

func (s *sessionsStub) Profile(context.Context, string) (auth.Profile, error) {
	if s.revoked {
		return auth.Profile{}, auth.ErrUnauthorized
	}
	return auth.Profile{
		ID:       "profile",
		Nickname: "Guitarist",
		Email:    "guitarist@example.test",
	}, nil
}

func (s *sessionsStub) Revoke(context.Context, string) error { s.revoked = true; return nil }

func TestProfileContractAndCookieRevocation(t *testing.T) {
	sessions := &sessionsStub{}
	handler := New(sessions, nil, CookieConfig{SecureCookies: true})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.AddCookie(&http.Cookie{
		Name:  sessionCookieName,
		Value: "token",
	})
	response := httptest.NewRecorder()
	handler.Me(response, req)
	var profile map[string]json.RawMessage
	if err := json.Unmarshal(response.Body.Bytes(), &profile); err != nil {
		t.Fatal(err)
	}
	if response.Code != http.StatusOK || profile["nickname"] == nil || profile["createdAt"] == nil || string(profile["email"]) != `"guitarist@example.test"` {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	response = httptest.NewRecorder()
	handler.Logout(response, req)
	cookies := response.Result().Cookies()
	if response.Code != http.StatusNoContent || !sessions.revoked || len(cookies) != 1 || cookies[0].MaxAge != -1 || !cookies[0].Secure || !cookies[0].HttpOnly {
		t.Fatalf("logout status=%d cookies=%v", response.Code, cookies)
	}
}

func TestInternalValidationRequiresServiceSecret(t *testing.T) {
	h := New(nil, nil, CookieConfig{InternalSecret: "secret"})
	req := httptest.NewRequest(http.MethodPost, "/internal/sessions/validate", nil)
	response := httptest.NewRecorder()
	h.Validate(response, req)
	if response.Code != http.StatusForbidden {
		t.Fatalf("status=%d", response.Code)
	}
}
