package managementapi

import (
	"context"

	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/internal/user"
	"authcore.io/authcore/pkg/api/managementapi"
	"authcore.io/authcore/pkg/nulls"
)

// GetMetadata gets the user metadata for a given user.
func (s *Service) GetMetadata(ctx context.Context, in *managementapi.GetMetadataRequest) (*managementapi.GetMetadataResponse, error) {
	currentUser, ok := user.CurrentUserFromContext(ctx)
	if !ok || currentUser == nil {
		return nil, errors.New(errors.ErrorUnauthenticated, "")
	}

	err := s.authorize(ctx, GetMetadataPermission)
	if err != nil {
		return nil, err
	}

	user, err := s.UserStore.UserByPublicID(ctx, in.UserId)
	if err != nil {
		return nil, err
	}
	userMetadata, err := user.UserMetadata.String()
	if err != nil {
		return nil, err
	}
	appMetadata, err := user.AppMetadata.String()
	if err != nil {
		return nil, err
	}

	return &managementapi.GetMetadataResponse{
		UserMetadata: userMetadata,
		AppMetadata:  appMetadata,
	}, nil
}

// UpdateMetadata updates the user metadata for a given user.
func (s *Service) UpdateMetadata(ctx context.Context, in *managementapi.UpdateMetadataRequest) (*managementapi.UpdateMetadataResponse, error) {
	currentUser, ok := user.CurrentUserFromContext(ctx)
	if !ok || currentUser == nil {
		return nil, errors.New(errors.ErrorUnauthenticated, "")
	}

	err := s.authorize(ctx, UpdateMetadataPermission)
	if err != nil {
		return nil, err
	}

	u, err := s.UserStore.UserByPublicID(ctx, in.UserId)
	if err != nil {
		return nil, err
	}

	if in.UserMetadata != "" {
		userMetadata := nulls.NewJSON(in.UserMetadata)
		u.UserMetadata = userMetadata
	}
	if in.AppMetadata != "" {
		appMetadata := nulls.NewJSON(in.AppMetadata)
		u.AppMetadata = appMetadata
	}
	err = s.UserStore.UpdateUser(ctx, u)
	if err != nil {
		return nil, err
	}

	updatedUserMetadata, err := u.UserMetadata.String()
	if err != nil {
		return nil, err
	}
	updatedAppMetadata, err := u.AppMetadata.String()
	if err != nil {
		return nil, err
	}

	return &managementapi.UpdateMetadataResponse{
		UserMetadata: updatedUserMetadata,
		AppMetadata:  updatedAppMetadata,
	}, nil
}
