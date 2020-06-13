package authapi

import (
	"context"
	"strings"

	"authcore.io/authcore/internal/clientapp"
	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/pkg/api/authapi"
	"authcore.io/authcore/pkg/httputil"
)

// ValidateOAuthParameters validates the parameters for OAuth authorization request.
// TODO: Validate the parameter properly when (1) the concept of "app" is developed, (2) further
// response type suport, (3) support multiple domains and (4) for "plain" code challenge
func (s *Service) ValidateOAuthParameters(ctx context.Context, in *authapi.ValidateOAuthParametersRequest) (*authapi.ValidateOAuthParametersResponse, error) {
	clientID := in.ClientId
	clientApp, err := clientapp.GetByClientID(clientID)
	if clientApp == nil {
		return nil, errors.New(errors.ErrorInvalidArgument, "")
	}
	if in.ResponseType != "code" {
		return nil, errors.New(errors.ErrorInvalidArgument, "")
	}
	if in.CodeChallenge != "" {
		if in.CodeChallengeMethod != "S256" {
			return nil, errors.New(errors.ErrorInvalidArgument, "")
		}
	}
	acceptURIPrefixes := clientApp.AllowedCallbackURLs
	err = validateRedirectURI(in.RedirectUri, acceptURIPrefixes)
	if err != nil {
		return nil, err
	}
	return &authapi.ValidateOAuthParametersResponse{}, nil
}

func validateRedirectURI(uri string, acceptURIPrefixes []string) error {
	normalizedURI, err := httputil.NormalizeURI(uri)
	if err != nil {
		return errors.Wrap(err, errors.ErrorInvalidArgument, "")
	}
	for _, acceptURIPrefix := range acceptURIPrefixes {
		if strings.HasPrefix(normalizedURI, acceptURIPrefix) {
			return nil
		}
	}
	return errors.New(errors.ErrorInvalidArgument, "")
}
