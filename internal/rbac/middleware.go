package rbac

import (
	"authcore.io/authcore/internal/errors"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	log "github.com/sirupsen/logrus"
)

const (
	subjectKey = "subject"
	guestKey   = "guest"
)

// EnforcerMiddleware returns a middleware to authorize requests with casbin.
func EnforcerMiddleware(skipper middleware.Skipper, enforcer *Enforcer) echo.MiddlewareFunc {
	if skipper == nil {
		skipper = middleware.DefaultSkipper
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if skipper(c) {
				return next(c)
			}
			sub := c.Get(subjectKey)
			if sub == nil {
				sub = guestKey
			}
			obj := c.Path()
			act := c.Request().Method
			allowed, err := enforcer.Enforce(sub, obj, act)
			if err != nil {
				return errors.Wrap(err, errors.ErrorUnknown, "")
			}

			if !allowed {
				roles, err := enforcer.GetRolesForUser(sub.(string))
				if err != nil {
					roles = []string{}
				}
				log.WithFields(log.Fields{
					"sub":   sub,
					"obj":   obj,
					"act":   act,
					"roles": roles,
				}).Error("authorization denied")
				if sub == guestKey {
					return errors.New(errors.ErrorUnauthenticated, "authentication is required")
				}
				return errors.New(errors.ErrorPermissionDenied, "")
			}

			return next(c)
		}
	}
}
