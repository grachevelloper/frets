package auth

import (
	"crypto/sha256"
	"time"
)

const SessionLifetime = 30 * 24 * time.Hour

func HashSecret(value string) []byte {
	sum := sha256.Sum256([]byte(value))
	return sum[:]
}
