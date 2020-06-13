package verifier

import (
	"testing"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/hotp"
	"github.com/stretchr/testify/assert"
)

func TestBackupCodeVerifier(t *testing.T) {
	f := NewFactory()

	data := `{
		"method": "backup_code",
		"secret": "THISISASECRETFORBACKUPCODETESTSX",
		"used_code_mask": "31"
	}`
	verifier, err := f.Unmarshal([]byte(data))
	assert.NoError(t, err)
	assert.Equal(t, "backup_code", verifier.Method())
	assert.False(t, verifier.IsPrimary())
	assert.False(t, verifier.SkipMFA())
	assert.Empty(t, verifier.Salt())
	backupCodeVerifier, ok := verifier.(BackupCodeVerifier)
	assert.True(t, ok)
	assert.Equal(t, int64(31), backupCodeVerifier.UsedCodeMask)

	code, err := hotp.GenerateCodeCustom("THISISASECRETFORBACKUPCODETESTSX", uint64(6), hotp.ValidateOpts{
		Digits:    otp.DigitsEight,
		Algorithm: otp.AlgorithmSHA1,
	})
	ok, newVerifier := verifier.Verify([]byte{}, []byte(code))
	assert.True(t, ok)
	assert.Equal(t, int64(95), newVerifier.(BackupCodeVerifier).UsedCodeMask)
}
