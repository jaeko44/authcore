package authentication

import (
	"context"

	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/pkg/cryptoutil"

	"github.com/google/uuid"
	"github.com/spf13/viper"
)

const proofOfWorkChallengeKeyPrefix = "challenge/pow/"

// CreateProofOfWorkChallenge creates a new proof-of-work challenge.
func (srv *Service) CreateProofOfWorkChallenge(ctx context.Context, intent string) (*ProofOfWorkChallenge, error) {
	expiry := viper.GetDuration("pow_challenge_time_limit")

	token := uuid.New().String()
	challenge, err := cryptoutil.CreateProofOfWorkChallenge(intent)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	err = srv.Redis.Set(proofOfWorkChallengeKeyPrefix+token, challenge, expiry).Err()
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	powChallenge := &ProofOfWorkChallenge{
		Token:      token,
		Challenge:  []byte(challenge),
		Difficulty: int64(viper.GetInt("pow_challenge_difficulty")),
		// Purpose: ...,
	}

	err = powChallenge.Validate()
	if err != nil {
		return nil, err
	}

	return powChallenge, nil
}

// BurnProofOfWorkResponse verifies Proof-of-work response and remove the challenges from Redis to prevent replay.
func (srv *Service) BurnProofOfWorkResponse(ctx context.Context, challengeToken string, proof []byte) error {
	// Verify PoW
	challenge, err := srv.Redis.Get(proofOfWorkChallengeKeyPrefix + challengeToken).Result()
	if err != nil {
		return errors.Wrap(err, errors.ErrorNotFound, "")
	}

	difficulty := int64(viper.GetInt("pow_challenge_difficulty"))
	_, err = cryptoutil.VerifyProofOfWork(challenge, string(proof), difficulty)
	if err != nil {
		// Proof of work is not valid
		return errors.Wrap(err, errors.ErrorInvalidArgument, "")
	}

	srv.Redis.Del(proofOfWorkChallengeKeyPrefix + challengeToken)
	return nil
}
