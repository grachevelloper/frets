package token

import (
	"crypto/rand"
	"encoding/base64"
)

type Generator struct{}

func (Generator) Generate() string {
	var data [32]byte
	rand.Read(data[:])
	return base64.RawURLEncoding.EncodeToString(data[:])
}
