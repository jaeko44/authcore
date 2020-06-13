package authentication

import (
	"encoding/json"

	"authcore.io/authcore/internal/errors"

	"github.com/go-redis/redis"
	"gitlab.com/blocksq/spake2-go"
)

// SPAKE2Data represent a SPAKE2+ key confirmation state to be stored during the SPAKE2+ response.
type SPAKE2Data struct {
	UserID             string
	Confirmation       []byte
	RemoteConfirmation []byte
}

const spake2DataKeyPrefix = "spake2data/"

// NewSPAKE2Plus initialize a SPAKE2Plus, with scrypt parameter now as 16384, 8, 1.
func NewSPAKE2Plus() (*spake2.SPAKE2Plus, error) {
	suite := spake2.Ed25519Sha256HkdfHmacScrypt(spake2.Scrypt(16384, 8, 1))
	spake2plus, err := spake2.NewSPAKE2Plus(suite)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	return spake2plus, nil
}

// GetIdentity gets the identity parameters for SPAKE2Plus, which is hardcoded as (idA, idB) = (authcoreuser, authcore)
func GetIdentity() (clientIdentity, serverIdentity []byte) {
	// for now it is hardcode idA and idB
	clientIdentity = []byte("authcoreuser")
	serverIdentity = []byte("authcore")
	return
}

// newSPAKE2Server initialize a SPAKE2 ServerPlusState for the SPAKE2+ response.
func (s *Service) newSPAKE2Server(userID string, verifierW0, verifierL []byte) (*spake2.ServerPlusState, []byte, error) {
	spake2Plus, err := NewSPAKE2Plus()
	if err != nil {
		return nil, nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	// now aad is nil, TODO: make it as hostname for mutual hostname agreement
	clientIdentity, serverIdentity := GetIdentity()
	state, msg, err := spake2Plus.StartServer(clientIdentity, serverIdentity, verifierW0, verifierL, nil)
	if err != nil {
		return nil, nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	return state, msg, nil
}

// retrieveSPAKE2Confirmations retrieves SPAKE2+ confirmations and re-construct a *spake2.Confirmations.
func (s *Service) retrieveSPAKE2Confirmations(challengeToken, userID string) (*spake2.Confirmations, error) {
	suite := spake2.Ed25519Sha256HkdfHmacScrypt(spake2.Scrypt(16384, 8, 1))

	jsonSPAKE2Data, err := s.Redis.Get(spake2DataKeyPrefix + challengeToken).Result()
	if err == redis.Nil {
		return nil, errors.New(errors.ErrorNotFound, "")
	} else if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	spake2data := &SPAKE2Data{}
	err = json.Unmarshal([]byte(jsonSPAKE2Data), spake2data)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	if spake2data.UserID != userID {
		return nil, errors.New(errors.ErrorUnauthenticated, "")
	}

	return spake2.NewConfirmations(spake2data.Confirmation, spake2data.RemoteConfirmation, suite), nil
}
