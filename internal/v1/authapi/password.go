package authapi

import (
	"context"

	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/internal/user"
	"authcore.io/authcore/pkg/api/authapi"
)

// StartChangePassword creates a change password flow and returns password parameters.
func (s *Service) StartChangePassword(ctx context.Context, in *authapi.StartChangePasswordRequest) (*authapi.StartChangePasswordResponse, error) {
	user, ok := user.CurrentUserFromContext(ctx)
	if !ok || user == nil {
		return nil, errors.New(errors.ErrorUnauthenticated, "")
	}

	return &authapi.StartChangePasswordResponse{
		Salt: user.PasswordSalt(),
	}, nil
}

// ChangePasswordKeyExchange creates a password challenge.
func (s *Service) ChangePasswordKeyExchange(ctx context.Context, in *authapi.ChangePasswordKeyExchangeRequest) (*authapi.PasswordChallenge, error) {
	user, ok := user.CurrentUserFromContext(ctx)
	if !ok || user == nil {
		return nil, errors.New(errors.ErrorUnauthenticated, "")
	}

	challenge, err := s.AuthenticationService.NewPasswordChallengeWithUser(ctx, user, in.Message)
	if err != nil {
		return nil, err
	}

	return challenge, nil
}

// FinishChangePassword updates user password. The client should requests a PasswordChallenge with CreatePasswordChallenge
// prior to calling ChangePassword if the user password was set.
func (s *Service) FinishChangePassword(ctx context.Context, in *authapi.FinishChangePasswordRequest) (*authapi.FinishChangePasswordResponse, error) {
	if in.PasswordVerifier == nil || len(in.PasswordVerifier.Salt) == 0 || len(in.PasswordVerifier.VerifierW0) == 0 || len(in.PasswordVerifier.VerifierL) == 0 {
		return nil, errors.New(errors.ErrorInvalidArgument, "")
	}

	user, ok := user.CurrentUserFromContext(ctx)
	if !ok || user == nil {
		return nil, errors.New(errors.ErrorUnauthenticated, "")
	}

	// If password is set, requires verifying SPAKE2 challenge
	if user.IsPasswordAuthenticationEnabled() {
		passwordResponse := in.OldPasswordResponse
		if passwordResponse == nil {
			return nil, errors.New(errors.ErrorInvalidArgument, "")
		}

		err := s.AuthenticationService.VerifyPasswordResponseWithUser(ctx, user, passwordResponse.Token, passwordResponse.Confirmation)
		if err != nil {
			return nil, errors.New(errors.ErrorPermissionDenied, "")
		}
	}

	err := user.SetPasswordVerifier(in.PasswordVerifier.Salt, in.PasswordVerifier.VerifierW0, in.PasswordVerifier.VerifierL)
	if err != nil {
		return nil, errors.New(errors.ErrorInvalidArgument, "")
	}

	err = s.UserStore.UpdateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	return &authapi.FinishChangePasswordResponse{}, nil
}

// ResetPassword resets the users' password if proper reset password token is given.  A client should obtain the reset password token
// from AuthenticateResetPassword flow.
func (s *Service) ResetPassword(ctx context.Context, in *authapi.ResetPasswordRequest) (*authapi.ResetPasswordResponse, error) {
	if in.PasswordVerifier == nil || len(in.PasswordVerifier.Salt) == 0 || len(in.PasswordVerifier.VerifierW0) == 0 || len(in.PasswordVerifier.VerifierL) == 0 {
		return nil, errors.New(errors.ErrorInvalidArgument, "")
	}

	token := in.Token
	resetPasswordToken, err := s.AuthenticationService.FindResetPasswordToken(ctx, token)
	if err != nil {
		return nil, err
	}

	user, err := s.UserStore.UserByID(ctx, resetPasswordToken.UserID)
	if err != nil {
		return nil, err
	}

	user, err = s.UserStore.ClearResetPasswordCount(ctx, user)
	if err != nil {
		return nil, err
	}

	err = user.SetPasswordVerifier(in.PasswordVerifier.Salt, in.PasswordVerifier.VerifierW0, in.PasswordVerifier.VerifierL)
	if err != nil {
		return nil, errors.New(errors.ErrorInvalidArgument, "")
	}

	err = s.UserStore.UpdateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	err = s.AuthenticationService.DeleteResetPasswordToken(ctx, token)
	if err != nil {
		return nil, err
	}

	return &authapi.ResetPasswordResponse{}, nil
}
