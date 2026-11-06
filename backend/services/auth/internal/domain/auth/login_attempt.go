package auth

import "time"

const LoginLifetime = 10 * time.Minute

type Attempt struct {
	StateHash   []byte
	BrowserHash []byte
	Provider    string
	Verifier    string
	ReturnTo    string
	ExpiresAt   time.Time
}
