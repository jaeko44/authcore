package authentication

import (
	"context"
	"testing"

	"authcore.io/authcore/pkg/cryptoutil"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestCreateProofOfWorkChallenge(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()
	viper.SetDefault("pow_challenge_difficulty", "4294967296")

	challenge, err := srv.CreateProofOfWorkChallenge(context.Background(), "")
	if assert.Nil(t, err) {
		assert.NotNil(t, challenge)
		assert.Equal(t, challenge.Difficulty, int64(4294967296))
	}
}

func TestBurnProofOfWorkResponse(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()
	viper.SetDefault("pow_challenge_difficulty", "1024")

	challenge, err := srv.CreateProofOfWorkChallenge(context.TODO(), "")
	if assert.Nil(t, err) {
		assert.NotNil(t, challenge)
		assert.NotZero(t, challenge.Difficulty)
	}

	proof, err := cryptoutil.SolveProofOfWork(string(challenge.Challenge), challenge.Difficulty)
	assert.Nil(t, err)

	err = srv.BurnProofOfWorkResponse(context.Background(), challenge.Token, []byte(proof))
	assert.Nil(t, err)
}
