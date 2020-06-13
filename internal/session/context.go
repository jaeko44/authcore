package session

import (
	"context"

	"github.com/labstack/echo/v4"
)

type currentsessionKey struct{}

// NewContextWithCurrentSession constructs a context with the current session.
func NewContextWithCurrentSession(ctx context.Context, session *Session) context.Context {
	return context.WithValue(ctx, currentsessionKey{}, session)
}

// CurrentSessionFromContext returns the current session in ctx if it exists. The returned *Session should not be modified.
func CurrentSessionFromContext(ctx context.Context) (session *Session, ok bool) {
	session, ok = ctx.Value(currentsessionKey{}).(*Session)
	return
}

// FromContext returns the current session saved in an echo.Context.
func FromContext(c echo.Context) (sess *Session, ok bool) {
	sess, ok = c.Get(sessionKey).(*Session)
	return
}
