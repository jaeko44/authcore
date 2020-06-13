package session

import (
	"crypto/ecdsa"
	"strings"
	"time"

	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/internal/user"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/spf13/viper"
)

// AccessToken represents a token bundle with an access token, ID Token and an expiry time.
type AccessToken struct {
	AccessToken string
	IDToken     string
	ExpiresIn   int64
}

func generateAccessToken(signer *ecdsa.PrivateKey, userID, sessionID, audience string, userRecord *user.User) (AccessToken, error) {
	expiresIn := viper.GetDuration("access_token_expires_in")
	issuer := viper.GetString("base_url")
	issuedAt := time.Now()
	expireAt := issuedAt.Add(expiresIn)

	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"iat": issuedAt.Unix(),
		"exp": expireAt.Unix(),
		"iss": issuer,
		"sub": userID,
		"sid": sessionID,
		"aud": audience,
	})
	keyID, err := kidFromECPublicKey(&signer.PublicKey)
	if err != nil {
		return AccessToken{}, err
	}
	token.Header["kid"] = keyID

	tokenString, err := token.SignedString(signer)
	if err != nil {
		return AccessToken{}, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	var idTokenString string
	if userRecord != nil {
		idToken := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
			"iat":                   issuedAt.Unix(),
			"exp":                   expireAt.Unix(),
			"iss":                   issuer,
			"sub":                   userID,
			"sid":                   sessionID,
			"aud":                   audience,
			"name":                  userRecord.DisplayName(),
			"email":                 userRecord.Email.String,
			"email_verified":        userRecord.EmailVerifiedAt.Valid,
			"phone_number":          userRecord.Phone.String,
			"phone_number_verified": userRecord.PhoneVerifiedAt.Valid,
			"preferred_username":    userRecord.Username.String,
		})

		idTokenString, err = idToken.SignedString(signer)
		if err != nil {
			return AccessToken{}, errors.Wrap(err, errors.ErrorUnknown, "")
		}
	}

	return AccessToken{
		AccessToken: tokenString,
		IDToken:     idTokenString,
		ExpiresIn:   int64(expiresIn.Seconds()),
	}, nil
}

// verifyAccessToken verifies the signature of a give JWT token and returns the asserted userID and sessionID.
func verifyAccessToken(accessTokenPublicKey *ecdsa.PublicKey, serviceAccountsMap map[string]ServiceAccount, token string) (userID string, sessionID string, err error) {
	jwtToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if method, ok := token.Method.(*jwt.SigningMethodECDSA); !ok || method.Alg() != "ES256" {
			return nil, errors.New(errors.ErrorInvalidArgument, "")
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return nil, errors.New(errors.ErrorInvalidArgument, "")
		}

		issuer, ok := claims["iss"].(string)
		if ok && strings.HasPrefix(issuer, ServiceAccountPrefix) {
			return serviceAccountKeyFunc(serviceAccountsMap, token)
		}
		kid, err := kidFromECPublicKey(accessTokenPublicKey)
		if err != nil {
			return nil, err
		}
		if token.Header["kid"] != kid {
			return nil, errors.Errorf(errors.ErrorUnauthenticated, "unrecognized kid: %v", token.Header["kid"])
		}
		return accessTokenPublicKey, nil
	})
	if err != nil {
		return "", "", errors.Wrap(err, errors.ErrorUnauthenticated, "")
	}
	if !jwtToken.Valid {
		return "", "", errors.New(errors.ErrorInvalidArgument, "")
	}
	claims, ok := jwtToken.Claims.(jwt.MapClaims)
	if !ok {
		return "", "", errors.New(errors.ErrorInvalidArgument, "")
	}
	err = claims.Valid()
	if err != nil {
		return "", "", errors.New(errors.ErrorInvalidArgument, "")
	}

	sub, ok := claims["sub"].(string)
	if !ok {
		return "", "", errors.New(errors.ErrorInvalidArgument, "")
	}
	sid, ok := claims["sid"].(string)
	if !ok {
		sid = ""
	}
	return sub, sid, nil
}

func serviceAccountKeyFunc(serviceAccountsMap map[string]ServiceAccount, token *jwt.Token) (*ecdsa.PublicKey, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New(errors.ErrorInvalidArgument, "")
	}
	issuer, ok := claims["iss"].(string)
	if !ok {
		return nil, errors.New(errors.ErrorInvalidArgument, "")
	}
	sub, ok := claims["sub"].(string)
	if !ok {
		return nil, errors.New(errors.ErrorInvalidArgument, "")
	}
	if sub != issuer {
		return nil, errors.Errorf(errors.ErrorUnauthenticated, "unexpected subject: %v", sub)
	}

	id := issuer[len(ServiceAccountPrefix):]
	sa, ok := serviceAccountsMap[id]
	if !ok {
		return nil, errors.Errorf(errors.ErrorUnauthenticated, "unrecognized service account: %v", issuer)
	}
	return sa.PublicKey()
}
