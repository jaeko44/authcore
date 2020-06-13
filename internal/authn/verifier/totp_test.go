package verifier

import (
	"testing"
	"time"

	"authcore.io/authcore/pkg/cryptoutil"

	"github.com/stretchr/testify/assert"
)

func TestTOTPVerifier(t *testing.T) {
	f := NewFactory()

	data := `{
		"method": "totp",
		"secret": "THISISATOTPSECRETXXXXXXXXXXXXXXX"
	}`
	verifier, err := f.Unmarshal([]byte(data))
	assert.NoError(t, err)
	_, ok := verifier.(TOTPVerifier)
	assert.True(t, ok)
	assert.Equal(t, "totp", verifier.Method())
	assert.False(t, verifier.IsPrimary())
	assert.False(t, verifier.SkipMFA())
	assert.Empty(t, verifier.Salt())
	code := cryptoutil.GetTOTPPin("THISISATOTPSECRETXXXXXXXXXXXXXXX", time.Now())
	ok, newVerifier := verifier.Verify([]byte{}, []byte(code))
	assert.True(t, ok)
	assert.NotEqual(t, int64(0), newVerifier.(TOTPVerifier).LastUsed)
}
