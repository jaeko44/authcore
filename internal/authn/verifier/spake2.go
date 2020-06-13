package verifier

import (
	"crypto/rand"
	"encoding/json"

	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/pkg/slice"

	log "github.com/sirupsen/logrus"
	"gitlab.com/blocksq/spake2-go"
)

const (
	// SPAKE2Plus represents SPAKE2+ algorithm.
	SPAKE2Plus string = "spake2plus"
)

// SPAKE2PlusVerifier implements SPAKE2+ PAKE.
type SPAKE2PlusVerifier struct {
	MethodName string `json:"method"`
	SaltValue  []byte `json:"salt"`
	W0         []byte `json:"w0"`
	L          []byte `json:"l"`
}

// NewSPAKE2PlusVerifier returns a new SPAKE2PlusVerifier instance.
func NewSPAKE2PlusVerifier(salt, w0, l []byte) SPAKE2PlusVerifier {
	return SPAKE2PlusVerifier{
		MethodName: SPAKE2Plus,
		SaltValue:  salt,
		W0:         w0,
		L:          l,
	}
}

// Method returns "spake2plus".
func (v SPAKE2PlusVerifier) Method() string {
	return SPAKE2Plus
}

// IsPrimary returns true as this method can be used as a primary factor.
func (v SPAKE2PlusVerifier) IsPrimary() bool {
	return true
}

// SkipMFA returns whether this method is sufficient for completing the authentication.
func (v SPAKE2PlusVerifier) SkipMFA() bool {
	return false
}

// Salt returns a password salt
func (v SPAKE2PlusVerifier) Salt() []byte {
	return v.SaltValue
}

// Request generates a new password challenge and a state for completing the verification later.
func (v SPAKE2PlusVerifier) Request(in []byte) (state State, challenge Challenge, err error) {
	if len(v.W0) == 0 || len(v.L) == 0 {
		log.Warnf("spake2plus: invalid verifier")
		err = errors.New(errors.ErrorInvalidArgument, "invalid verifier")
		return
	}
	if len(in) == 0 || slice.IsZeroBytes(in) {
		log.Warnf("spake2plus: empty key exchange message")
		err = errors.New(errors.ErrorInvalidArgument, "invalid key exchange message")
		return
	}
	clientIdentity := []byte("authcoreuser")
	serverIdentity := []byte("authcore")

	spake2plus, err := NewSPAKE2Plus()
	if err != nil {
		return
	}
	s, m, err := spake2plus.StartServer(clientIdentity, serverIdentity, v.W0, v.L, nil)
	if err != nil {
		return
	}

	sk, err := s.Finish(in)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Warn("spake2plus: key exchange failed")
		err = errors.Wrap(err, errors.ErrorUnauthenticated, "")
		return
	}

	_, state = sk.GetConfirmations()
	challenge = m

	return
}

// Verify verifies the incoming password response
func (v SPAKE2PlusVerifier) Verify(state State, in []byte) (bool, Verifier) {
	if len(state) != 32 {
		log.Warnf("spake2plus: attempt to verify password without key exchange")
		return false, nil
	}
	if len(in) != 32 {
		log.Warnf("spake2plus: attempt to verify password with invalid confirmation")
		return false, nil
	}
	suite := spake2.Ed25519Sha256HkdfHmacScrypt(spake2.Scrypt(16384, 8, 1))
	return suite.MacEqual(in, state), nil
}

// GenerateSPAKE2Verifier generates a SPAKE2 verifier from plaintext password. It is for unit tests and creating first admin account.
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

// SPAKE2PlusVerifierFromJSON unmarshals a SPAKE2PlusVerifier from a JSON data.
func SPAKE2PlusVerifierFromJSON(data []byte) (v Verifier, err error) {
	t := SPAKE2PlusVerifier{}
	if err = json.Unmarshal(data, &t); err != nil {
		return
	}
	v = t
	return
}

// generateSalt is a utility function that generate a cryptographic random salt with given length.
func generateSalt(len uint) ([]byte, error) {
	buffer := make([]byte, len)
	_, err := rand.Read(buffer)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	return buffer, nil
}

// NewSPAKE2Plus returns a SPAKE2Plus instance used in verifier.
func NewSPAKE2Plus() (spake2plus *spake2.SPAKE2Plus, err error) {
	suite := spake2.Ed25519Sha256HkdfHmacScrypt(spake2.Scrypt(16384, 8, 1))
	spake2plus, err = spake2.NewSPAKE2Plus(suite)
	return
}
