package secret

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/pkg/errors"
)

// String represents a string secret.
type String struct {
	string
}

// NewString returns a new SecretValue with the given secret.
func NewString(src string) String {
	return String{src}
}

// EmptyString returns an empty secret string.
func EmptyString() String {
	return String{}
}

// String returns SHA256('SHA256(secret):' + secret) that is safe for display.
func (s String) String() string {
	if s.string == "" {
		return ""
	}
	prefixed := "SHA256(secret):" + s.string
	hash := sha256.Sum256([]byte(prefixed))
	return "SHA256(secret):" + hex.EncodeToString(hash[:])
}

// MarshalJSON returns the masked secret as a JSON string.
func (s String) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%v\"", s)), nil
}

// SecretString returns the unmasked value of the secret.
func (s String) SecretString() string {
	return s.string
}

// SecretBytes16 decodes the secret as 32 characters hex string (128-bit) and returns bytes value, otherwise returns an
// error if the value is malformed or the length is not 32 characters.
func (s String) SecretBytes16() (bytes16 [16]byte, err error) {
	bytes, err := s.decodeBytes()
	if err != nil {
		return
	}

	if len(bytes) != 16 {
		err = errors.Errorf("secret is not 16 bytes")
		return
	}

	copy(bytes16[:], bytes[:16])
	return
}

// SecretBytes32 decodes the secret as 64 characters hex string (256-bit) and returns bytes value, otherwise returns an
// error if the value is malformed or the length is not 64 characters.
func (s String) SecretBytes32() (bytes32 [32]byte, err error) {
	bytes, err := s.decodeBytes()
	if err != nil {
		return
	}

	if len(bytes) != 32 {
		err = errors.Errorf("secret is not 32 bytes")
		return
	}

	copy(bytes32[:], bytes[:32])
	return
}

// SecretBytes decodes the secret as a hex string and returns bytes value, otherwise returns an
// error if the value is malformed.
func (s String) SecretBytes() (bytes []byte, err error) {
	bytes, err = s.decodeBytes()
	return
}

func (s String) decodeBytes() ([]byte, error) {
	bytes, err := hex.DecodeString(s.string)
	if err != nil {
		return nil, errors.Wrap(err, "invalid hex value")
	}
	return bytes, nil
}
