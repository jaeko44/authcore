package authn

import (
	"context"
	"testing"

	"authcore.io/authcore/internal/config"
	"authcore.io/authcore/internal/testutil"
	"authcore.io/authcore/pkg/cryptoutil"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func storeForTest() (*Store, func()) {
	config.InitDefaults()
	viper.Set("secret_key_base", "855edf399835e9c9deb61877c1a76bf14eed7c35a167e10ff1b7d43db4363268")
	config.InitConfig()
	redis := testutil.RedisForTest()
	encryptor := testutil.EncryptorForTest()
	store := NewStore(redis, encryptor)

	return store, func() {
		viper.Reset()
		redis.FlushAll()
	}
}

func TestPutState(t *testing.T) {
	s, teardown := storeForTest()
	defer teardown()

	state := &State{
		ClientID:            "app",
		StateToken:          cryptoutil.RandomToken32(),
		Status:              StatusPrimary,
		UserID:              1,
		Factors:             []string{FactorPassword},
		RedirectURI:         "https://example.com/",
		PKCEChallengeMethod: "code_challenge_method",
		PKCEChallenge:       "code_challenge",
	}
	ctx := context.Background()
	err := s.PutState(ctx, state)
	assert.NoError(t, err)

	state.ClearFactors()

	state2, err := s.GetState(ctx, state.StateToken)
	assert.NoError(t, err)
	assert.Equal(t, state, state2)
}

func TestPutAuthorizationCode(t *testing.T) {
	s, teardown := storeForTest()
	defer teardown()

	state := &State{
		ClientID:            "app",
		StateToken:          cryptoutil.RandomToken32(),
		Status:              StatusPrimary,
		UserID:              1,
		Factors:             []string{FactorPassword},
		RedirectURI:         "https://example.com/",
		PKCEChallengeMethod: "code_challenge_method",
		PKCEChallenge:       "code_challenge",
	}
	ctx := context.Background()
	err := s.PutState(ctx, state)
	assert.NoError(t, err)

	code := state.GenerateAuthorizationCode()
	err = s.PutAuthorizationCode(ctx, code)
	assert.NoError(t, err)
	assert.Equal(t, state.AuthorizationCode, code.Code)

	code2, err := s.GetAuthorizationCode(ctx, code.Code)
	assert.NoError(t, err)
	assert.Equal(t, code, code2)
}
