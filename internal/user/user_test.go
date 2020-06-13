package user

import (
	"context"
	"testing"

	"authcore.io/authcore/internal/authn/verifier"
	"github.com/stretchr/testify/assert"
)

func TestPasswordVerifierFromUserNoPassword(t *testing.T) {
	store, teardown := storeForTest()
	defer teardown()

	ctx := context.Background()
	u, err := store.UserByID(ctx, 1)
	if !assert.NoError(t, err) {
		return
	}

	_, err = u.PasswordVerifier()
	assert.Error(t, err)
}

func TestFromSecondFactor(t *testing.T) {
	store, teardown := storeForTest()
	defer teardown()

	ctx := context.Background()
	f := verifier.NewFactory()

	totpFactor, err := store.FindSecondFactorByID(ctx, 2)
	if !assert.NoError(t, err) {
		return
	}
	v, err := totpFactor.ToVerifier(f)
	assert.NoError(t, err)
	totpVerifier, ok := v.(verifier.TOTPVerifier)
	assert.True(t, ok)
	assert.Equal(t, "totp", totpVerifier.MethodName)
	assert.Equal(t, "THISISAWEAKTOTPSECRETFORTESTSXX2", totpVerifier.Secret)
	assert.Equal(t, int64(1546992000), totpVerifier.LastUsed)

	backupCodeFactor, err := store.FindSecondFactorByID(ctx, 6)
	if !assert.NoError(t, err) {
		return
	}
	v, err = backupCodeFactor.ToVerifier(f)
	assert.NoError(t, err)
	backupCodeVerifier, ok := v.(verifier.BackupCodeVerifier)
	assert.True(t, ok)
	assert.Equal(t, "backup_code", backupCodeVerifier.MethodName)
	assert.Equal(t, "THISISASECRETFORBACKUPCODETESTSX", backupCodeVerifier.Secret)
	assert.Equal(t, int64(31), backupCodeVerifier.UsedCodeMask)
}

func TestUpdateSecondFactorContentWithVerifier(t *testing.T) {
	store, teardown := storeForTest()
	defer teardown()

	ctx := context.Background()
	f := verifier.NewFactory()

	totpFactor, err := store.FindSecondFactorByID(ctx, 6)
	if !assert.NoError(t, err) {
		return
	}
	v, err := totpFactor.ToVerifier(f)
	assert.NoError(t, err)
	backupCodeVerifier, ok := v.(verifier.BackupCodeVerifier)
	assert.True(t, ok)
	assert.Equal(t, "backup_code", backupCodeVerifier.MethodName)
	assert.Equal(t, "THISISASECRETFORBACKUPCODETESTSX", backupCodeVerifier.Secret)
	assert.Equal(t, int64(31), backupCodeVerifier.UsedCodeMask)

	backupCodeVerifier.UsedCodeMask = 33
	totpFactor.UpdateWithVerifier(backupCodeVerifier)

	assert.True(t, totpFactor.Content.UsedCodeMask.Valid)
	assert.Equal(t, int64(33), totpFactor.Content.UsedCodeMask.Int64)
}
