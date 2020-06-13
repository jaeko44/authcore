package audit

import (
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

// Auditor represents an audit service.
type Auditor interface {
	LogEvent(c echo.Context, actor Actor, action string, target interface{}, result EventResult)
}

// NewLoggingAuditor returns an auditor that write events to log. Usually for testing.
func NewLoggingAuditor() Auditor {
	return &auditor{}
}

// auditor is an auditor that write events to log.
type auditor struct {
}

func (a *auditor) LogEvent(c echo.Context, actor Actor, action string, target interface{}, result EventResult) {
	logrus.WithFields(logrus.Fields{
		"actor_id": actor.ActorID(),
		"actor":    actor.DisplayName(),
		"action":   action,
		"target":   target,
		"result":   result.String(),
	}).Info("audit event")
}
