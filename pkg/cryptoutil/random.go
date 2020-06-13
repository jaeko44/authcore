package cryptoutil

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/base64"
	"fmt"
	"math/big"

	log "github.com/sirupsen/logrus"
)

// RandomToken returns a 16 bytes cryptographically secure random token encoded in URL-safe base64 no padding. Panics if
// it fail to generate the token.
func RandomToken() string {
	token := make([]byte, 16)
	_, err := rand.Read(token)
	if err != nil {
		log.Fatal("failed to generate random token")
	}
	return base64.RawURLEncoding.EncodeToString(token)
}

// RandomToken32 returns a 32 bytes cryptographically secure random token encoded in URL-safe base64 no padding. Panics if
// it fail to generate the token.
func RandomToken32() string {
	token := make([]byte, 32)
	_, err := rand.Read(token)
	if err != nil {
		log.Fatal("failed to generate random token")
	}
	return base64.RawURLEncoding.EncodeToString(token)
}

// RandomCode returns a n-digit code. Panics if it fails to generate the code.
func RandomCode(digits int64) string {
	maxValue := new(big.Int).Exp(big.NewInt(10), big.NewInt(digits), nil)
	code, err := rand.Int(rand.Reader, maxValue)
	if err != nil {
		log.Fatal("failed to generate random code")
	}
	template := fmt.Sprintf("%%0%dv", digits)
	paddedCode := fmt.Sprintf(template, code.String())
	return paddedCode
}

// RandomBackupCodeSecret returns a 32-character base32 string as a secret for backup code.
// (1) 20-byte entropy is generated according to section 4 of RFC4226 (https://tools.ietf.org/html/rfc4226),
//     which "RECOMMENDs a shared secret length of 160 bits".
// (2) Base32 is commonly used to store secrets for HOTP / TOTP.
//     For instance, in the OTP package used (https://godoc.org/github.com/pquerna/otp),
//     `ErrValidateSecretInvalidBase32` error is returned if base32 is not used.
func RandomBackupCodeSecret() string {
	backupCodeSecret := make([]byte, 20)
	_, err := rand.Read(backupCodeSecret)
	if err != nil {
		log.Fatal("failed to generate random token")
	}
	return base32.StdEncoding.EncodeToString(backupCodeSecret)
}
