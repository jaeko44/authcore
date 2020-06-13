package session

import (
	"crypto/sha256"
	"encoding/base64"
)

func computeRefreshTokenHash(token string) string {
	hash := sha256.Sum256([]byte(token))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}
