package server

import (
	"context"
	"strings"

	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/internal/session"
	"authcore.io/authcore/internal/user"
	"authcore.io/authcore/pkg/api/authapi"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

const authorizationKey = "authorization"
const bearerKey = "Bearer"

// AccessTokenResolver provides a hook to verify an access token and return the asserted userID.
type AccessTokenResolver func(ctx context.Context, token string) (userID, sessionID string, err error)

// UserResolver provides a hook to return a *db.User object corresponding to the userID.
type UserResolver func(ctx context.Context, userID string) (user *user.User, err error)

// SessionResolver provides a hook to return a *db.Session object corresponding to the credential.
type SessionResolver func(ctx context.Context, sessionID string) (user *session.Session, err error)

// NewAuthorizationUnaryInterceptor implements grpc.UnaryServerInterceptor to authorize requests with bearer token.
func NewAuthorizationUnaryInterceptor(accessTokenResolver AccessTokenResolver, userResolver UserResolver, sessionResolver SessionResolver) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			headers := md.Get(authorizationKey)
			if len(headers) > 0 {
				parts := strings.SplitN(headers[0], " ", 2)
				if len(parts) == 2 && parts[0] == bearerKey {
					token := strings.TrimSpace(parts[1])
					userID, sessionID, err := accessTokenResolver(ctx, token)
					if err != nil {
						return nil, err
					}
					u, err := userResolver(ctx, userID)
					if err != nil {
						return nil, errors.Wrap(err, errors.ErrorPermissionDenied, "")
					}
					ctx = user.NewContextWithCurrentUser(ctx, u)

					if sessionID != "" {
						sess, err := sessionResolver(ctx, sessionID)
						if err != nil {
							return nil, errors.Wrap(err, errors.ErrorPermissionDenied, "")
						}
						if sess.UserID != u.ID {
							return nil, errors.New(errors.ErrorPermissionDenied, "")
						}
						ctx = session.NewContextWithCurrentSession(ctx, sess)

						log.WithFields(log.Fields{
							"user_id": userID,
							"error":   err,
						}).Info("Access token authentication success")
					} else {
						log.WithFields(log.Fields{
							"user_id": userID,
						}).Info("Service account authentication success")
					}

				}
			}
		}

		return handler(ctx, req)
	}
}

// tokenAccess supplies PerRPCCredentials from a given token.
type tokenAccess struct {
	token *authapi.AccessToken
}

// NewTokenAccess constructs the PerRPCCredentials using a given token.
func NewTokenAccess(token *authapi.AccessToken) credentials.PerRPCCredentials {
	return tokenAccess{token: token}
}

func (ta tokenAccess) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": "Bearer" + " " + ta.token.AccessToken,
	}, nil
}

func (ta tokenAccess) RequireTransportSecurity() bool {
	return false
}
