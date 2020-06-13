package cryptoutil

import (
	"crypto/sha256"
	"encoding/base64"
	"math/big"

	"github.com/pkg/errors"
)

// CreateProofOfWorkChallenge generates a proof of work challenge and return the token and the challenge
func CreateProofOfWorkChallenge(_ string) (string, error) {
	return RandomToken(), nil
}

// VerifyProofOfWork verifies whether a proof of work response is correct for the challenge with given difficulty
func VerifyProofOfWork(challenge string, response string, difficulty int64) (bool, error) {
	// SHA256(challenge || response) < 2^256 / difficulty
	bChallenge, err := base64.RawURLEncoding.DecodeString(challenge)
	if err != nil {
		return false, errors.Wrap(err, "failed to decode challenge with base 64")
	}
	if len(bChallenge) != 16 {
		err := errors.New("the challenge is not of 16 bytes long")
		return false, err
	}
	bResponse, err := base64.RawURLEncoding.DecodeString(response)
	if err != nil {
		return false, errors.Wrap(err, "failed to decode response with base 64")
	}
	if len(bResponse) != 16 {
		err := errors.New("the response is not of 16 bytes long")
		return false, err
	}

	var hashValue big.Int
	hashBytes := sha256.Sum256(append(bChallenge, bResponse...))
	hash := hashValue.SetBytes(hashBytes[:])

	var powTargetValue big.Int
	powTarget := powTargetValue.Div(powTargetValue.Lsh(big.NewInt(1), 256), big.NewInt(difficulty))

	if hash.Cmp(powTarget) >= 0 {
		err := errors.New("invalid proof of work")
		return false, err
	}
	return true, nil
}

// SolveProofOfWork computes a proof of work response for the given challenge and difficulty.
func SolveProofOfWork(challenge string, difficulty int64) (string, error) {
	var powTargetValue big.Int
	powTarget := powTargetValue.Div(powTargetValue.Lsh(big.NewInt(1), 256), big.NewInt(difficulty))
	bChallenge, err := base64.RawURLEncoding.DecodeString(string(challenge))
	if err != nil {
		return "", errors.Wrap(err, "failed to decode challenge with base 64")
	}

	proof := make([]byte, 16)
	intProof := big.NewInt(0)
	for {
		bProof := intProof.Bytes()
		copy(proof[16-len(bProof):], bProof)

		var hashValue big.Int
		hashBytes := sha256.Sum256(append(bChallenge, proof...))
		hash := hashValue.SetBytes(hashBytes[:])
		if hash.Cmp(powTarget) < 0 {
			return base64.RawURLEncoding.EncodeToString(proof), nil
		}
		intProof.Add(intProof, big.NewInt(1))
	}
}
