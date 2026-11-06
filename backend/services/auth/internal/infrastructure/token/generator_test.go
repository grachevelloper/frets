package token

import (
	"encoding/base64"
	"testing"
)

func TestGenerator(t *testing.T) {
	generator := Generator{}
	seen := make(map[string]bool)
	for range 100 {
		token := generator.Generate()
		decoded, err := base64.RawURLEncoding.DecodeString(token)
		if err != nil || len(decoded) != 32 || seen[token] {
			t.Fatalf("invalid or repeated token: length=%d error=%v", len(decoded), err)
		}
		seen[token] = true
	}
}
