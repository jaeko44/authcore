package cryptoutil

import (
	"github.com/pkg/errors"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/hotp"
	log "github.com/sirupsen/logrus"
)

// ValidateBackupCodes validates if a passcode, with counters being 0, 1, 2, ... or limit-1
// (excluding the masked numbers), is correct.
func ValidateBackupCodes(passcode string, secret string, usedCodeMask int64, limit uint64) (int64, error) {
	for counter := uint64(0); counter < limit; counter++ {
		if usedCodeMask&(1<<counter) != 0 {
			continue
		}
		if ValidateBackupCode(passcode, secret, counter) {
			return usedCodeMask | (1 << counter), nil
		}
	}
	return int64(0), errors.New("invalid passcode")
}

// ValidateBackupCode validates if a passcode is correct.
func ValidateBackupCode(passcode string, secret string, counter uint64) bool {
	valid, err := hotp.ValidateCustom(passcode, counter, secret, hotp.ValidateOpts{
		Digits:    otp.DigitsEight,
		Algorithm: otp.AlgorithmSHA1,
	})
	return err == nil && valid
}

// GetBackupCodePin gets the BackupCode PIN for given secret and counter.
func GetBackupCodePin(secret string, counter uint64) string {
	code, err := hotp.GenerateCodeCustom(secret, counter, hotp.ValidateOpts{
		Digits:    otp.DigitsEight,
		Algorithm: otp.AlgorithmSHA1,
	})
	if err != nil {
		log.Panic("cannot generate code")
	}
	return code
}
