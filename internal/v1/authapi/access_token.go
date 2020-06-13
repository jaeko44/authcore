package authapi

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/base64"
	"sync"

	"authcore.io/authcore/internal/audit"
	"authcore.io/authcore/internal/db"
	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/pkg/api/authapi"
	"authcore.io/authcore/pkg/secret"

	jwt "github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var parsePrivateKeysOnce sync.Once

// AccessTokenPrivateKey is the private key for signing access token.
var AccessTokenPrivateKey *ecdsa.PrivateKey

// AccessTokenPublicKey is the public key for verifying access token.
var AccessTokenPublicKey *ecdsa.PublicKey

// CreateAccessToken validates a session and returns a new short-term JWT access token.
func (s *Service) CreateAccessToken(ctx context.Context, in *authapi.CreateAccessTokenRequest) (*authapi.AccessToken, error) {
	token := in.Token
	grantType := in.GrantType
	codeVerifier := in.CodeVerifier

	var accessToken *authapi.AccessToken
	var err error
	switch grantType {
	case authapi.CreateAccessTokenRequest_AUTHORIZATION_TOKEN:
		accessToken, err = s.createAccessTokenByAuthorizationToken(ctx, token, codeVerifier)
	case authapi.CreateAccessTokenRequest_REFRESH_TOKEN:
		accessToken, err = s.createAccessTokenByRefreshToken(ctx, token)
	}

	if err != nil {
		return nil, err
	}
	return accessToken, nil
}

func (s *Service) createAccessTokenByAuthorizationToken(ctx context.Context, sAuthorizationToken string, codeVerifier string) (*authapi.AccessToken, error) {
	authorizationToken, err := s.AuthenticationService.FindAuthorizationToken(ctx, sAuthorizationToken)
	if errors.IsKind(err, errors.ErrorNotFound) {
		return nil, errors.Wrap(err, errors.ErrorPermissionDenied, "")
	} else if err != nil {
		return nil, err
	}

	user, err := s.UserStore.UserByID(ctx, authorizationToken.UserID)
	if errors.IsKind(err, errors.ErrorNotFound) {
		return nil, errors.Wrap(err, errors.ErrorPermissionDenied, "")
	} else if err != nil {
		return nil, err
	}

	if user.IsCurrentlyLocked() {
		return nil, errors.New(errors.ErrorPermissionDenied, "user is locked")
	}

	if !isCodeChallengeValid(codeVerifier, authorizationToken.CodeChallenge, "S256") {
		return nil, errors.New(errors.ErrorInvalidArgument, "")
	}
	err = s.AuthenticationService.DeleteAuthorizationToken(ctx, sAuthorizationToken)
	if err != nil {
		return nil, err
	}

	session, err := s.SessionStore.CreateSession(ctx, authorizationToken.UserID, authorizationToken.DeviceID, authorizationToken.ClientID, "", false)
	if err != nil {
		return nil, err
	}
	log.WithFields(log.Fields{
		"user_id":    user.PublicID(),
		"session_id": session.PublicID(),
	}).Info("session created for user")
	s.AuditStore.CreateEvent(ctx, audit.SystemActor, "authcore.user.authenticated", db.NullableInt64(session.ID), user, audit.EventResultSuccess)

	refreshToken := session.Refresh(ctx, true)
	s.SessionStore.UpdateSession(ctx, session)

	bundle, err := s.SessionStore.GenerateAccessToken(ctx, session, true)
	if err != nil {
		return nil, err
	}

	return &authapi.AccessToken{
		AccessToken:  bundle.AccessToken,
		IdToken:      bundle.IDToken,
		RefreshToken: refreshToken,
		TokenType:    authapi.AccessToken_BEARER,
		ExpiresIn:    bundle.ExpiresIn,
	}, nil
}

func (s *Service) createAccessTokenByRefreshToken(ctx context.Context, sRefreshToken string) (*authapi.AccessToken, error) {
	session, err := s.SessionStore.FindSessionByRefreshToken(ctx, sRefreshToken)
	if errors.IsKind(err, errors.ErrorNotFound) {
		return nil, errors.Wrap(err, errors.ErrorPermissionDenied, "")
	} else if err != nil {
		return nil, err
	}

	user, err := s.UserStore.UserByID(ctx, session.UserID)
	if errors.IsKind(err, errors.ErrorNotFound) {
		return nil, errors.Wrap(err, errors.ErrorPermissionDenied, "")
	} else if err != nil {
		return nil, err
	}

	if user.IsCurrentlyLocked() {
		return nil, errors.New(errors.ErrorPermissionDenied, "user is locked")
	}

	refreshToken := session.Refresh(ctx, false)
	s.SessionStore.UpdateSession(ctx, session)

	bundle, err := s.SessionStore.GenerateAccessToken(ctx, session, true)
	if err != nil {
		return nil, err
	}

	return &authapi.AccessToken{
		AccessToken:  bundle.AccessToken,
		IdToken:      bundle.IDToken,
		RefreshToken: refreshToken,
		TokenType:    authapi.AccessToken_BEARER,
		ExpiresIn:    bundle.ExpiresIn,
	}, nil
}

func parsePrivateKeys() {
	parsePrivateKeysOnce.Do(func() {
		privateKeyPEM := viper.Get("access_token_private_key").(secret.String).SecretString()
		if privateKeyPEM != "" {
			privateKey, err := jwt.ParseECPrivateKeyFromPEM([]byte(privateKeyPEM))
			if err != nil {
				log.Fatalf("invalid access_token_private_key: %v", err)
			}
			AccessTokenPrivateKey = privateKey
			AccessTokenPublicKey = &privateKey.PublicKey
		}
	})
}

func isCodeChallengeValid(codeVerifier, codeChallenge, codeChallengeMethod string) bool {
	// if challenge is not set, then assume that PKCE is not enabled
	if codeChallenge == "" {
		return true
	}
	switch codeChallengeMethod {
	case "plain":
		return false // TODO: Implement the "plain" code challenge method
	case "S256":
		verifier := []byte(codeVerifier)
		challenge, err := base64.RawURLEncoding.DecodeString(codeChallenge)
		if err != nil {
			return false
		}
		hashVerifier := sha256.Sum256(verifier)
		return bytes.Compare(hashVerifier[:], challenge) == 0
	default:
		return false
	}
}
