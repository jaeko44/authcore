package authapi

import (
	"context"

	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/internal/user"
	"authcore.io/authcore/pkg/api/authapi"
	"authcore.io/authcore/pkg/nulls"
)

// GetMetadata gets the user metadata for the current user
func (s *Service) GetMetadata(ctx context.Context, in *authapi.GetMetadataRequest) (*authapi.GetMetadataResponse, error) {
	currentUser, ok := user.CurrentUserFromContext(ctx)
	if !ok || currentUser == nil {
		return nil, errors.New(errors.ErrorUnauthenticated, "")
	}

	userMetadata, err := currentUser.UserMetadata.String()
	if err != nil {
		return nil, err
	}
	return &authapi.GetMetadataResponse{
		UserMetadata: userMetadata,
	}, nil
}

// UpdateMetadata updates the user metadata for the current user
func (s *Service) UpdateMetadata(ctx context.Context, in *authapi.UpdateMetadataRequest) (*authapi.UpdateMetadataResponse, error) {
	currentUser, ok := user.CurrentUserFromContext(ctx)
	if !ok || currentUser == nil {
		return nil, errors.New(errors.ErrorUnauthenticated, "")
	}

	userMetadata := nulls.NewJSON(in.UserMetadata)
	currentUser.UserMetadata = userMetadata
	err := s.UserStore.UpdateUser(ctx, currentUser)
	if err != nil {
		return nil, err
	}
	updatedUserMetadata, err := userMetadata.String()
	if err != nil {
		return nil, err
	}

	return &authapi.UpdateMetadataResponse{
		UserMetadata: updatedUserMetadata,
	}, nil
}
