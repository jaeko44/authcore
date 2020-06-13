package messageencryptor

import (
	"crypto/sha256"
	"io"
	"strings"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/hkdf"
)

// KeyGenerator represents a key generator, which generates a key from a secret and salt
// with HKDF-SHA256.
type KeyGenerator struct {
	prk []byte // pseudorandom key for expand
}

// NewKeyGenerator returns a new key generator.
func NewKeyGenerator(secret []byte) *KeyGenerator {
	hash := sha256.New
	salt := []byte(strings.Repeat("\x00", 32))
	return &KeyGenerator{
		prk: hkdf.Extract(hash, secret, salt),
	}
}

// Derive derives a key, of given length from secret and info, for cryptographic uses.
func (keyGenerator *KeyGenerator) Derive(info string, length int) []byte {
	hash := sha256.New
	hkdfStream := hkdf.Expand(hash, keyGenerator.prk, []byte(info))
	key := make([]byte, length)
	_, err := io.ReadFull(hkdfStream, key)
	if err != nil {
		log.WithFields(log.Fields{
			"error":  err,
			"info":   info,
			"length": length,
		}).Fatal("cannot derive key from key generator")
	}
	return key
}
