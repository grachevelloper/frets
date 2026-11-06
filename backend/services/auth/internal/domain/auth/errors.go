package auth

import "errors"

var (
	ErrUnauthorized        = errors.New("unauthorized")
	ErrUnknownProvider     = errors.New("not_found")
	ErrProviderUnavailable = errors.New("provider_unavailable")
	ErrLoginInvalid        = errors.New("login_invalid")
	ErrLoginFailed         = errors.New("login_failed")
)
