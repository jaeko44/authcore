package managementapi

import (
	"context"
	"strconv"

	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/internal/user"
	"authcore.io/authcore/pkg/api/authapi"
	"authcore.io/authcore/pkg/api/managementapi"

	"github.com/golang/protobuf/ptypes/timestamp"
)

// ListOAuthFactors lists the OAuth factors of a user.
func (s *Service) ListOAuthFactors(ctx context.Context, in *managementapi.ListOAuthFactorsRequest) (*managementapi.ListOAuthFactorsResponse, error) {
	currentUser, ok := user.CurrentUserFromContext(ctx)
	if !ok || currentUser == nil {
		return nil, errors.New(errors.ErrorUnauthenticated, "")
	}
	err := s.authorize(ctx, ListOAuthFactorsPermission)
	if err != nil {
		return nil, err
	}

	userID, err := strconv.ParseInt(in.UserId, 10, 64)

	oauthFactors, err := s.UserStore.FindAllOAuthFactorsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var pbOAuthFactors []*authapi.OAuthFactor
	for _, oauthFactor := range *oauthFactors {
		pbOAuthFactor, err := MarshalOAuthFactor(&oauthFactor)
		if err != nil {
			return nil, err
		}
		pbOAuthFactors = append(pbOAuthFactors, pbOAuthFactor)
	}
	return &managementapi.ListOAuthFactorsResponse{
		OauthFactors: pbOAuthFactors,
	}, nil
}

// DeleteOAuthFactor deletes an OAuth factor (of some user).
func (s *Service) DeleteOAuthFactor(ctx context.Context, in *managementapi.DeleteOAuthFactorRequest) (*managementapi.DeleteOAuthFactorResponse, error) {
	currentUser, ok := user.CurrentUserFromContext(ctx)
	if !ok || currentUser == nil {
		return nil, errors.New(errors.ErrorUnauthenticated, "")
	}

	err := s.authorize(ctx, DeleteOAuthFactorsPermission)
	if err != nil {
		return nil, err
	}

	id, err := strconv.ParseInt(in.Id, 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	oauthFactor, err := s.UserStore.FindOAuthFactorByID(ctx, id)
	if err != nil {
		return nil, err
	}
	user, err := s.UserStore.UserByID(ctx, oauthFactor.UserID)
	if err != nil {
		return nil, err
	}

	// Check if the OAuth factor is the only one and password is set
	oauthFactors, err := s.UserStore.FindAllOAuthFactorsByUserID(ctx, oauthFactor.UserID)
	if err != nil {
		return nil, err
	}
	if len(*oauthFactors) == 1 {
		if !user.IsPasswordAuthenticationEnabled() {
			return nil, errors.New(errors.ErrorInvalidArgument, "")
		}
	}

	// The authenticator is not belong to the current user
	if oauthFactor == nil {
		return nil, errors.New(errors.ErrorPermissionDenied, "")
	}

	err = s.UserStore.DeleteOAuthFactorByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &managementapi.DeleteOAuthFactorResponse{}, nil
}

// MarshalOAuthFactor marshals a *user.OAuthFactor into Protobuf message
func MarshalOAuthFactor(in *user.OAuthFactor) (*authapi.OAuthFactor, error) {
	metadata, err := in.Metadata.String()
	if err != nil {
		return nil, err
	}
	return &authapi.OAuthFactor{
		Id:          in.ID,
		UserId:      in.UserID,
		Service:     authapi.OAuthFactor_OAuthService(in.Service),
		OauthUserId: in.OAuthUserID,
		CreatedAt: &timestamp.Timestamp{
			Seconds: in.CreatedAt.Unix(),
			Nanos:   int32(in.CreatedAt.Nanosecond()),
		},
		LastUsedAt: &timestamp.Timestamp{
			Seconds: in.LastUsedAt.Unix(),
			Nanos:   int32(in.LastUsedAt.Nanosecond()),
		},
		Metadata: metadata,
	}, nil
}
