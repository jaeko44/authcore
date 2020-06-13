package cryptoutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetTOTPPin(t *testing.T) {
	secret := "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	totpPin := GetTOTPPin(secret, time.Now())

	assert.Equal(t, len(totpPin), 6)
}

func TestValidateTOTP(t *testing.T) {
	d, _ := time.ParseDuration("-24h")
	lastUsedAt := time.Now().Add(d)

	authenticator := &TOTPAuthenticator{
		TotpSecret: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		LastUsedAt: lastUsedAt,
	}
	totpPin := GetTOTPPin(authenticator.TotpSecret, time.Now())

	assert.True(t, ValidateTOTP(totpPin, authenticator))
}

func TestValidateTOTPWithLeeway(t *testing.T) {
	d, _ := time.ParseDuration("-24h")
	d2, _ := time.ParseDuration("-30s")
	lastUsedAt := time.Now().Add(d)

	authenticator := &TOTPAuthenticator{
		TotpSecret: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		LastUsedAt: lastUsedAt,
	}
	totpPin := GetTOTPPin(authenticator.TotpSecret, time.Now().Add(d2))

	assert.True(t, ValidateTOTP(totpPin, authenticator))
}

func TestValidateTOTPWrongPin(t *testing.T) {
	d, _ := time.ParseDuration("-24h")
	lastUsedAt := time.Now().Add(d)

	authenticator := &TOTPAuthenticator{
		TotpSecret: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		LastUsedAt: lastUsedAt,
	}
	totpPin := "000000"

	assert.False(t, ValidateTOTP(totpPin, authenticator))
}

func TestValidateTOTPRecentlyUsed(t *testing.T) {
	lastUsedAt := time.Now()

	authenticator := &TOTPAuthenticator{
		TotpSecret: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		LastUsedAt: lastUsedAt,
	}
	totpPin := GetTOTPPin(authenticator.TotpSecret, time.Now())

	assert.False(t, ValidateTOTP(totpPin, authenticator))
}

func TestValidateTOTPs(t *testing.T) {
	d, _ := time.ParseDuration("-24h")
	lastUsedAt := time.Now().Add(d)

	authenticators := []*TOTPAuthenticator{}
	// Add some authenticators
	authenticators = append(authenticators, &TOTPAuthenticator{
		TotpSecret: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		LastUsedAt: lastUsedAt,
	})
	authenticators = append(authenticators, &TOTPAuthenticator{
		TotpSecret: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAB",
		LastUsedAt: lastUsedAt,
	})
	authenticators = append(authenticators, &TOTPAuthenticator{
		TotpSecret: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAC",
		LastUsedAt: lastUsedAt,
	})

	totpPin := GetTOTPPin(authenticators[0].TotpSecret, time.Now())
	res, err := ValidateTOTPs(totpPin, authenticators)
	assert.Nil(t, err)
	assert.Equal(t, 0, res)

	totpPin2 := GetTOTPPin(authenticators[1].TotpSecret, time.Now())
	res2, err := ValidateTOTPs(totpPin2, authenticators)
	assert.Nil(t, err)
	assert.Equal(t, 1, res2)

	totpPin3 := GetTOTPPin(authenticators[2].TotpSecret, time.Now())
	res3, err := ValidateTOTPs(totpPin3, authenticators)
	assert.Nil(t, err)
	assert.Equal(t, 2, res3)
}

func TestValidateTOTPsWrongPin(t *testing.T) {
	d, _ := time.ParseDuration("-24h")
	lastUsedAt := time.Now().Add(d)

	authenticators := []*TOTPAuthenticator{}
	// Add some authenticators
	authenticators = append(authenticators, &TOTPAuthenticator{
		TotpSecret: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		LastUsedAt: lastUsedAt,
	})
	authenticators = append(authenticators, &TOTPAuthenticator{
		TotpSecret: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAB",
		LastUsedAt: lastUsedAt,
	})
	authenticators = append(authenticators, &TOTPAuthenticator{
		TotpSecret: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAC",
		LastUsedAt: lastUsedAt,
	})

	totpPin := "000000"
	_, err := ValidateTOTPs(totpPin, authenticators)
	assert.Error(t, err)
}
