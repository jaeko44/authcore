package verifier

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSPAKE2PlusVerifier(t *testing.T) {
	f := NewFactory()

	clientIdentity := []byte("authcoreuser")
	serverIdentity := []byte("authcore")

	data := `{
		"method": "spake2plus",
		"salt": "/Jb4pAuatq5rrwdNRGRqW+PhlqzNR1pYtp1N5YWEn7s=",
		"w0": "MK43qvO3DoflMLl1AlZPJA==",
		"l": "gGQMM78ixkqSUTxq3LvyLuURe1bXQSUbqCPsEHfg65M="
	}`

	verifier, err := f.Unmarshal([]byte(data))
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "spake2plus", verifier.Method())
	assert.True(t, verifier.IsPrimary())
	assert.False(t, verifier.SkipMFA())
	assert.NotEmpty(t, verifier.Salt())
	salt, err := base64.StdEncoding.DecodeString("/Jb4pAuatq5rrwdNRGRqW+PhlqzNR1pYtp1N5YWEn7s=")
	if !assert.NoError(t, err) {
		return
	}
	spake2, err := NewSPAKE2Plus()
	if !assert.NoError(t, err) {
		return
	}

	// Correct password
	cs, message, err := spake2.StartClient(clientIdentity, serverIdentity, []byte("password"), salt, nil)
	if !assert.NoError(t, err) {
		return
	}

	vs, challenge, err := verifier.Request(message)
	if !assert.NoError(t, err) {
		return
	}
	assert.NotEmpty(t, challenge)
	sk, err := cs.Finish(challenge)
	assert.NoError(t, err)
	confirmation := sk.GetConfirmation()
	ok, _ := verifier.Verify(vs, confirmation)
	assert.True(t, ok)

	// Incorrect password
	cs, message, err = spake2.StartClient(clientIdentity, serverIdentity, []byte("wrong_password"), salt, nil)
	if !assert.NoError(t, err) {
		return
	}
	vs, challenge, err = verifier.Request(message)
	if !assert.NoError(t, err) {
		return
	}
	assert.NotEmpty(t, challenge)
	sk, err = cs.Finish(challenge)
	assert.NoError(t, err)
	confirmation = sk.GetConfirmation()
	ok, _ = verifier.Verify(vs, confirmation)
	assert.False(t, ok)

	// Invalid verifier state
	ok, _ = verifier.Verify(nil, confirmation)
	assert.False(t, ok)
	ok, _ = verifier.Verify([]byte{}, confirmation)
	assert.False(t, ok)
	ok, _ = verifier.Verify(make([]byte, 32), confirmation)
	assert.False(t, ok)

	// Empty key exchange message
	_, _, err = verifier.Request([]byte{})
	assert.Error(t, err)
	_, _, err = verifier.Request(make([]byte, 32))
	assert.Error(t, err)
}

func TestSPAKE2PlusVerifierInvalid(t *testing.T) {
	f := NewFactory()

	clientIdentity := []byte("authcoreuser")
	serverIdentity := []byte("authcore")

	data := `{
		"method": "spake2plus",
		"salt": "/Jb4pAuatq5rrwdNRGRqW+PhlqzNR1pYtp1N5YWEn7s="
	}`

	verifier, err := f.Unmarshal([]byte(data))
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "spake2plus", verifier.Method())
	assert.True(t, verifier.IsPrimary())
	assert.False(t, verifier.SkipMFA())
	assert.NotEmpty(t, verifier.Salt())
	salt, err := base64.StdEncoding.DecodeString("/Jb4pAuatq5rrwdNRGRqW+PhlqzNR1pYtp1N5YWEn7s=")
	if !assert.NoError(t, err) {
		return
	}
	spake2, err := NewSPAKE2Plus()
	if !assert.NoError(t, err) {
		return
	}

	// Correct password
	_, message, err := spake2.StartClient(clientIdentity, serverIdentity, []byte("password"), salt, nil)
	if !assert.NoError(t, err) {
		return
	}
	_, _, err = verifier.Request(message)
	assert.Error(t, err)
}
