package cryptoutil

import (
	"math"
	"time"

	"github.com/pkg/errors"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/hotp"
	"github.com/pquerna/otp/totp"
	// log "github.com/sirupsen/logrus"
)

// TOTPAuthenticator represents a TOTP authenticator record for cryptoutil.
type TOTPAuthenticator struct {
	TotpSecret string
	LastUsedAt time.Time
}

// ValidateTOTPs validates if a passcode is correct for the given array of TOTP secrets.
func ValidateTOTPs(passcode string, authenticators []*TOTPAuthenticator) (int, error) {
	for index, authenticator := range authenticators {
		isValid := ValidateTOTP(passcode, authenticator)
		if isValid {
			return index, nil
		}
	}
	return -1, errors.New("passcode is incorrect")
}

// ValidateTOTP validates if a passcode is correct for the given TOTP secret.
func ValidateTOTP(passcode string, authenticator *TOTPAuthenticator) bool {
	t := time.Now()

	counterLeewayStart := math.Floor(float64(t.Unix())/float64(30) - 1)
	counterLastUsed := math.Floor(float64(authenticator.LastUsedAt.Unix())/float64(30) + 1)
	counterLeewayEnd := math.Floor(float64(t.Unix()) / float64(30))

	minCounter := uint64(math.Max(counterLeewayStart, counterLastUsed))
	maxCounter := uint64(counterLeewayEnd)

	for counter := maxCounter; counter >= minCounter; counter-- {
		rv, err := hotp.ValidateCustom(passcode, counter, authenticator.TotpSecret, hotp.ValidateOpts{
			Digits:    otp.DigitsSix,
			Algorithm: otp.AlgorithmSHA1,
		})
		if err == nil && rv {
			return true
		}
	}
	return false
}

// GetTOTPPin gets the current TOTP PIN for the given TOTP secret.
func GetTOTPPin(secret string, t time.Time) string {
	totpPin, err := totp.GenerateCode(secret, t)
	if err != nil {
		panic(err)
	}
	return totpPin
}
