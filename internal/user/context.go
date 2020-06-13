package user

import (
	"context"

	"github.com/labstack/echo/v4"
)

type currentuserKey struct{}
type currentsessionKey struct{}

const (
	userKey = "user"
	sessionKey = "session"
)

// NewContextWithCurrentUser constructs a context with the current user.
func NewContextWithCurrentUser(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, currentuserKey{}, user)
}

// CurrentUserFromContext returns the current user information in ctx if it exists. The returned *User should not be modified.
func CurrentUserFromContext(ctx context.Context) (user *User, ok bool) {
	user, ok = ctx.Value(currentuserKey{}).(*User)
	return
}

// FromContext returns the current user saved in an echo.Context.
func FromContext(c echo.Context) (user *User, ok bool) {
	user, ok = c.Get(userKey).(*User)
	return
}


// sessionFromContext returns the current session saved in an echo.Context.
func sessionFromContext(c echo.Context) (sess session, ok bool) {
	sess, ok = c.Get(sessionKey).(session)
	return
}