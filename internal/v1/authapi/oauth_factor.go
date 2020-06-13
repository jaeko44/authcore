package authapi

import (
	"context"
	"fmt"
	"strconv"

	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/internal/user"
	"authcore.io/authcore/internal/v1/authapi/oauth"
	"authcore.io/authcore/pkg/api/authapi"
	"authcore.io/authcore/pkg/cryptoutil"
	"authcore.io/authcore/pkg/nulls"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/spf13/viper"
)

// StartCreateOAuthFactor initiates the create OAuth factor flow.
func (s *Service) StartCreateOAuthFactor(ctx context.Context, in *authapi.StartCreateOAuthFactorRequest) (*authapi.StartCreateOAuthFactorResponse, error) {
	currentUser, ok := user.CurrentUserFromContext(ctx)
	if !ok || currentUser == nil {
		return nil, errors.New(errors.ErrorUnauthenticated, "")
	}

	oauthFactor := oauth.GetFactorDelegate(in.Service)
	expiry := viper.GetDuration("create_oauth_factor_state_expires_in")
	oauthStateString := cryptoutil.RandomToken32()
	startCreateOAuthFactorKey := fmt.Sprintf("start_create_oauth_factor/%s", oauthStateString)
	err := s.Redis.Set(startCreateOAuthFactorKey, currentUser.ID, expiry).Err()
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	oauthEndpointURI, requestToken, err := oauth.GetOAuthEndpointURI(oauthFactor, oauthStateString)
	if err != nil {
		return nil, err
	}
	if requestToken != "" {
		key := fmt.Sprintf("oauth1/request_token/%s", oauthStateString)
		err := s.Redis.Set(key, requestToken, expiry).Err()
		if err != nil {
			return nil, errors.Wrap(err, errors.ErrorUnknown, "")
		}
	}
	return &authapi.StartCreateOAuthFactorResponse{
		OauthEndpointUri: oauthEndpointURI,
		State:            oauthStateString,
	}, nil
}

// CreateOAuthFactor creates an OAuth factor for the current user.
func (s *Service) CreateOAuthFactor(ctx context.Context, in *authapi.CreateOAuthFactorRequest) (*authapi.CreateOAuthFactorResponse, error) {
	currentUser, ok := user.CurrentUserFromContext(ctx)
	if !ok || currentUser == nil {
		return nil, errors.New(errors.ErrorUnauthenticated, "")
	}

	startCreateOAuthFactorKey := fmt.Sprintf("start_create_oauth_factor/%s", in.State)
	sUserID, err := s.Redis.Get(startCreateOAuthFactorKey).Result()
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorNotFound, "")
	}
	userID, err := strconv.ParseInt(sUserID, 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	if userID != currentUser.ID {
		return nil, errors.New(errors.ErrorPermissionDenied, "")
	}

	oauthFactors, err := s.UserStore.FindAllOAuthFactorsByUserIDAndService(ctx, currentUser.ID, user.OAuthService(in.Service))
	if err != nil {
		return nil, err
	}
	if len(*oauthFactors) > 0 {
		return nil, errors.New(errors.ErrorAlreadyExists, "")
	}

	key := fmt.Sprintf("oauth1/request_token/%s", in.State)
	requestToken, err := s.Redis.Get(key).Result()
	if err != nil {
		requestToken = ""
	}
	oauthFactor := oauth.GetFactorDelegate(in.Service)
	accessToken, idToken, err := oauth.GetTokensByAuthorizationCode(oauthFactor, in.Code, requestToken)
	if err != nil {
		return nil, err
	}
	oauthUser, err := oauthFactor.GetUser(accessToken, idToken)
	if err != nil {
		return nil, err
	}
	_, err = s.UserStore.FindOAuthFactorByOAuthIdentity(ctx, user.OAuthService(in.Service), oauthUser.ID)
	if err != nil && !errors.IsKind(err, errors.ErrorNotFound) {
		return nil, err
	} else if err == nil {
		return nil, errors.New(errors.ErrorAlreadyExists, "")
	}
	_, err = s.UserStore.CreateOAuthFactor(ctx, currentUser.ID, user.OAuthService(in.Service), oauthUser.ID, nulls.NewJSON(oauthUser.Metadata))
	if err != nil {
		return nil, err
	}
	return &authapi.CreateOAuthFactorResponse{}, nil
}

// CreateOAuthFactorByAccessToken creates an OAuth factor for the current user, using an encrypted access token.
func (s *Service) CreateOAuthFactorByAccessToken(ctx context.Context, in *authapi.CreateOAuthFactorByAccessTokenRequest) (*authapi.CreateOAuthFactorByAccessTokenResponse, error) {
	currentUser, ok := user.CurrentUserFromContext(ctx)
	if !ok || currentUser == nil {
		return nil, errors.New(errors.ErrorUnauthenticated, "")
	}

	oauthFactor := oauth.GetFactorDelegate(in.Service)

	accessToken := in.AccessToken // TODO: Decrypt the access token
	idToken := in.IdToken         // TODO: Decrypt the id token
	oauthFactors, err := s.UserStore.FindAllOAuthFactorsByUserIDAndService(ctx, currentUser.ID, user.OAuthService(in.Service))
	if err != nil {
		return nil, err
	}
	if len(*oauthFactors) > 0 {
		return nil, errors.New(errors.ErrorAlreadyExists, "")
	}

	oauthUser, err := oauthFactor.GetUser(accessToken, idToken)
	if err != nil {
		return nil, err
	}
	_, err = s.UserStore.FindOAuthFactorByOAuthIdentity(ctx, user.OAuthService(in.Service), oauthUser.ID)
	if err != nil && !errors.IsKind(err, errors.ErrorNotFound) {
		return nil, err
	} else if err == nil {
		return nil, errors.New(errors.ErrorAlreadyExists, "")
	}
	_, err = s.UserStore.CreateOAuthFactor(ctx, currentUser.ID, user.OAuthService(in.Service), oauthUser.ID, nulls.NewJSON(oauthUser.Metadata))
	if err != nil {
		return nil, err
	}

	return &authapi.CreateOAuthFactorByAccessTokenResponse{}, nil
}

// ListOAuthFactors lists the OAuth factors for the current user.
func (s *Service) ListOAuthFactors(ctx context.Context, in *authapi.ListOAuthFactorsRequest) (*authapi.ListOAuthFactorsResponse, error) {
	currentUser, ok := user.CurrentUserFromContext(ctx)
	if !ok || currentUser == nil {
		return nil, errors.New(errors.ErrorUnauthenticated, "")
	}
	oauthFactors, err := s.UserStore.FindAllOAuthFactorsByUserID(ctx, currentUser.ID)
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
	return &authapi.ListOAuthFactorsResponse{
		OauthFactors: pbOAuthFactors,
	}, nil
}

// DeleteOAuthFactor deletes an OAuth factor for the current user.
func (s *Service) DeleteOAuthFactor(ctx context.Context, in *authapi.DeleteOAuthFactorRequest) (*authapi.DeleteOAuthFactorResponse, error) {
	currentUser, ok := user.CurrentUserFromContext(ctx)
	if !ok || currentUser == nil {
		return nil, errors.New(errors.ErrorUnauthenticated, "")
	}
	id, err := strconv.ParseInt(in.Id, 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	// Check if the OAuth factor is the only one and password is set
	oauthFactors, err := s.UserStore.FindAllOAuthFactorsByUserID(ctx, currentUser.ID)
	if err != nil {
		return nil, err
	}
	if len(*oauthFactors) == 1 {
		if !currentUser.IsPasswordAuthenticationEnabled() {
			return nil, errors.New(errors.ErrorInvalidArgument, "")
		}
	}
	oauthFactor, err := s.UserStore.FindOAuthFactorByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if matter is disabled
	isMatterDisabled := viper.GetBool("matters_unlink_disabled")
	if isMatterDisabled && authapi.OAuthFactor_OAuthService(oauthFactor.Service) == authapi.OAuthFactor_MATTERS {
		return nil, errors.New(errors.ErrorPermissionDenied, "")
	}

	// The authenticator is not belong to the current user
	if oauthFactor == nil || oauthFactor.UserID != currentUser.ID {
		return nil, errors.New(errors.ErrorPermissionDenied, "")
	}

	err = s.UserStore.DeleteOAuthFactorByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &authapi.DeleteOAuthFactorResponse{}, nil
}

// StartAuthenticateOAuth initiates a OAuth authentication for the sign in flow.
func (s *Service) StartAuthenticateOAuth(ctx context.Context, in *authapi.StartAuthenticateOAuthRequest) (*authapi.StartAuthenticateOAuthResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "removed from v1.0")
}

// AuthenticateOAuth provides OAuth responses to update the state of an authentication flow.
func (s *Service) AuthenticateOAuth(ctx context.Context, in *authapi.AuthenticateOAuthRequest) (*authapi.AuthenticateOAuthResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "removed from v1.0")
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
		LastUsedAt: &timestamp.Timestamp{
			Seconds: in.LastUsedAt.Unix(),
			Nanos:   int32(in.LastUsedAt.Nanosecond()),
		},
		Metadata: metadata,
	}, nil
}
