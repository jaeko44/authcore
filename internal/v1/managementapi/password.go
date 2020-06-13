package managementapi

import (
	"context"

	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/internal/user"
	"authcore.io/authcore/pkg/api/managementapi"
)

// ChangePassword updates user password.
func (s *Service) ChangePassword(ctx context.Context, in *managementapi.ChangePasswordRequest) (*managementapi.ChangePasswordResponse, error) {
	if in.PasswordVerifier == nil || len(in.PasswordVerifier.Salt) == 0 || len(in.PasswordVerifier.VerifierW0) == 0 || len(in.PasswordVerifier.VerifierL) == 0 {
		return nil, errors.New(errors.ErrorInvalidArgument, "")
	}

	currentUser, ok := user.CurrentUserFromContext(ctx)
	if !ok || currentUser == nil {
		return nil, errors.New(errors.ErrorUnauthenticated, "")
	}

	err := s.authorize(ctx, UpdatePasswordPermission)
	if err != nil {
		return nil, err
	}

	user, err := s.UserStore.UserByPublicID(ctx, in.UserId)
	if err != nil {
		return nil, err
	}

	err = user.SetPasswordVerifier(in.PasswordVerifier.Salt, in.PasswordVerifier.VerifierW0, in.PasswordVerifier.VerifierL)
	if err != nil {
		return nil, err
	}

	err = s.UserStore.UpdateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	return &managementapi.ChangePasswordResponse{}, nil
}
