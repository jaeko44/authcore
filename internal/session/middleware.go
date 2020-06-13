package session

import (
	"context"
	"fmt"
	"strings"

	"authcore.io/authcore/internal/errors"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/mssola/user_agent"
)

// UserAgentKey refers to key for context value with user agent
type UserAgentKey struct{}

// IPKey refers to key for context value with IP address
type IPKey struct{}

const (
	userKey    = "user"
	userIDKey  = "user_id"
	sessionKey = "session"
	subjectKey = "subject"
)

var errJWTMissing = errors.New(errors.ErrorUnauthenticated, "missing or malformed jwt")

// AccessTokenAuthMiddleware returns a access token auth middleware.
func AccessTokenAuthMiddleware(skipper middleware.Skipper, store *Store) echo.MiddlewareFunc {
	if skipper == nil {
		skipper = middleware.DefaultSkipper
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if skipper(c) {
				return next(c)
			}

			token, err := jwtFromHeader(c)
			if err != nil {
				return next(c)
			}
			ctx := c.Request().Context()
			userID, sessionID, err := store.VerifyAccessToken(ctx, token)
			if err != nil {
				return err
			}

			if strings.HasPrefix(userID, ServiceAccountPrefix) {
				c.Set(subjectKey, userID)
			} else {
				u, err := store.userStore.UserByPublicID(ctx, userID)
				if err != nil {
					if errors.IsKind(err, errors.ErrorNotFound) {
						return errors.Wrap(err, errors.ErrorUnauthenticated, "current user not found")
					}
					return err
				}
				s, err := store.FindSessionByPublicID(ctx, sessionID)
				if err != nil {
					if errors.IsKind(err, errors.ErrorNotFound) {
						return errors.Wrap(err, errors.ErrorUnauthenticated, "invalid session")
					}
					return err
				}
				c.Set(userKey, u)
				c.Set(userIDKey, u.PublicID())
				c.Set(sessionKey, s)
				sub := fmt.Sprintf("u:%v", u.ID)
				c.Set(subjectKey, sub)
			}

			return next(c)
		}
	}
}

// UserAgentMiddleware injects user agent in context inside HTTP request
func UserAgentMiddleware(skipper middleware.Skipper) echo.MiddlewareFunc {
	if skipper == nil {
		skipper = middleware.DefaultSkipper
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userAgent := user_agent.New(c.Request().UserAgent())
			// updatedContext provides user agent and IP value with corresponding key
			updatedContext := context.WithValue(c.Request().Context(), UserAgentKey{}, userAgent)
			updatedContext = context.WithValue(updatedContext, IPKey{}, c.RealIP())
			requestWithUserAgentIP := c.Request().WithContext(updatedContext)
			c.SetRequest(requestWithUserAgentIP)
			return next(c)
		}
	}
}

func jwtFromHeader(c echo.Context) (string, error) {
	authScheme := "Bearer"
	auth := c.Request().Header.Get(echo.HeaderAuthorization)
	l := len(authScheme)
	if len(auth) > l+1 && auth[:l] == authScheme {
		return auth[l+1:], nil
	}
	return "", errJWTMissing
}
