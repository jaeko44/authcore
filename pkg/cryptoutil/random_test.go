package cryptoutil

import (
	"encoding/base64"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandomToken(t *testing.T) {
	token1 := RandomToken()
	assert.Equal(t, base64.RawURLEncoding.DecodedLen(len(token1)), 16)

	token2 := RandomToken()
	assert.Equal(t, base64.RawURLEncoding.DecodedLen(len(token2)), 16)

	assert.NotEqual(t, token1, token2)
}

func TestRandomToken32(t *testing.T) {
	token1 := RandomToken32()
	assert.Equal(t, base64.RawURLEncoding.DecodedLen(len(token1)), 32)

	token2 := RandomToken32()
	assert.Equal(t, base64.RawURLEncoding.DecodedLen(len(token2)), 32)

	assert.NotEqual(t, token1, token2)
}

func TestRandomCode(t *testing.T) {
	// Generated codes are random
	code1 := RandomCode(10)
	code2 := RandomCode(10)
	assert.NotEqual(t, code1, code2)

	// Generated codes must be n-digits long and prepend with zero if it is not filled
	for i := 0; i < 100; i++ {
		code := RandomCode(10)
		assert.Regexp(t, regexp.MustCompile("^[0-9]{10}$"), code)
	}
}

func TestRandomBackupCodeSecret(t *testing.T) {
	backupCodeSecret1 := RandomBackupCodeSecret()
	assert.Len(t, backupCodeSecret1, 32)

	backupCodeSecret2 := RandomBackupCodeSecret()
	assert.Len(t, backupCodeSecret2, 32)

	assert.NotEqual(t, backupCodeSecret1, backupCodeSecret2)
}
