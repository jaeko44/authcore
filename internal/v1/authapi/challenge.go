package authapi

import (
	"context"

	"authcore.io/authcore/internal/v1/authentication"
	"authcore.io/authcore/pkg/api/authapi"

	"github.com/golang/protobuf/ptypes/empty"
)

// CreateProofOfWorkChallenge returns a new proof-of-work challenge
func (s *Service) CreateProofOfWorkChallenge(ctx context.Context, in *empty.Empty) (*authapi.ProofOfWorkChallenge, error) {
	challenge, err := s.AuthenticationService.CreateProofOfWorkChallenge(ctx, "device")
	if err != nil {
		return nil, err
	}
	pbChallenge, err := MarshalProofOfWorkChallenge(challenge)
	if err != nil {
		return nil, err
	}
	return pbChallenge, nil
}

// MarshalProofOfWorkChallenge marshals a *db.ProofOfWorkChallenge into a protobuf message.
func MarshalProofOfWorkChallenge(in *authentication.ProofOfWorkChallenge) (*authapi.ProofOfWorkChallenge, error) {
	return &authapi.ProofOfWorkChallenge{
		Token:      in.Token,
		Challenge:  string(in.Challenge),
		Difficulty: in.Difficulty,
	}, nil
}
