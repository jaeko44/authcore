package authentication

import (
	"context"
	"crypto/rand"
	"encoding/json"

	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/internal/user"
	"authcore.io/authcore/pkg/api/authapi"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gitlab.com/blocksq/spake2-go"
)

// generateSalt is a utility function that generate a cryptographic random salt with given length.
func generateSalt(len uint) ([]byte, error) {
	buffer := make([]byte, len)
	_, err := rand.Read(buffer)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	return buffer, nil
}

// VerifyPasswordResponseWithUser verify the confirmation produced by a client.
func (s *Service) VerifyPasswordResponseWithUser(ctx context.Context, user *user.User, token string, incomingConfirmation []byte) error {
	if !user.IsPasswordAuthenticationEnabled() {
		return errors.New(errors.ErrorInvalidArgument, "password is not set")
	}

	confirmation, err := s.retrieveSPAKE2Confirmations(token, user.PublicID())
	if err != nil {
		return err
	}
	err = confirmation.Verify(incomingConfirmation)
	if err != nil {
		return errors.Wrap(err, errors.ErrorUnauthenticated, "")
	}
	return nil
}

// NewPasswordChallengeWithUser constructs SPAKE2+ ServerState, generates a challenge and returns a PasswordChallenge message.
func (s *Service) NewPasswordChallengeWithUser(ctx context.Context, user *user.User, incomingMessage []byte) (*authapi.PasswordChallenge, error) {
	verifierW0 := user.PasswordVerifierW0.ByteSlice
	verifierL := user.PasswordVerifierL.ByteSlice

	if !user.IsPasswordAuthenticationEnabled() {
		return nil, errors.New(errors.ErrorInvalidArgument, "")
	}

	spakeServer, response, err := s.newSPAKE2Server(user.PublicID(), verifierW0, verifierL)
	if err != nil {
		return nil, err
	}

	secret, err := spakeServer.Finish(incomingMessage)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnauthenticated, "")
	}

	confirmation, remoteConfirmation := secret.GetConfirmations()

	// TODO encrypt the secret before saving to Redis
	spake2Data := SPAKE2Data{
		UserID:             user.PublicID(),
		Confirmation:       confirmation,
		RemoteConfirmation: remoteConfirmation,
	}

	jsonSPAKE2Data, err := json.Marshal(spake2Data)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	challengeToken := uuid.New().String()
	timeLimit := viper.GetDuration("spake2_time_limit")
	log.Printf("Saving SPAKE2Data %v with time limit %v", challengeToken, timeLimit)
	err = s.Redis.Set(spake2DataKeyPrefix+challengeToken, jsonSPAKE2Data, timeLimit).Err()
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	challenge := &authapi.PasswordChallenge{
		Token:   challengeToken,
		Message: response,
	}

	return challenge, nil
}

// GenerateSPAKE2Verifier generates a SPAKE2 verifier from plaintext password. It is for unit tests and create first management account.
func GenerateSPAKE2Verifier(password, clientIdentity, serverIdentity []byte, spake2 *spake2.SPAKE2Plus) (salt, verifierW0, verifierL []byte, err error) {
	salt, err = generateSalt(32)
	if err != nil {
		err = errors.Wrap(err, errors.ErrorUnknown, "")
		return
	}

	verifierW0, verifierL, err = spake2.ComputeVerifier(password, salt, clientIdentity, serverIdentity)
	if err != nil {
		err = errors.Wrap(err, errors.ErrorUnknown, "")
	}
	return
}
