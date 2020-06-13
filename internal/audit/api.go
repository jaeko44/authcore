package audit

import (
	"net/http"
	"time"

	"authcore.io/authcore/internal/apiutil"
	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/pkg/nulls"

	"github.com/jinzhu/copier"
	"github.com/labstack/echo/v4"
)

// APIv2 returns a function that registers API 2.0 endpoints with an Echo instance.
func APIv2(store *Store) func(e *echo.Echo) {
	return func(e *echo.Echo) {
		h := &handler{store: store}

		g := e.Group("/api/v2")
		g.GET("/audit_logs", h.ListAuditLogs)
	}
}

type handler struct {
	store *Store
}

// ListAuditLogs is a API handler for list audit logs.
func (h *handler) ListAuditLogs(c echo.Context) error {
	r := EventsQuery{}
	if err := c.Bind(&r); err != nil {
		return err
	}

	if err := c.Validate(&r); err != nil {
		return err
	}

	ctx := c.Request().Context()

	events, page, err := h.store.AllEventsWithQuery(ctx, r)
	if err != nil {
		return err
	}
	jsonEvents := make([]JSONEvent, len(*events))
	for i, e := range *events {
		jsonEvents[i], err = NewJSONEvent(&e)
		if err != nil {
			return err
		}
	}
	resp := apiutil.NewListPagination(jsonEvents, page)
	return c.JSON(http.StatusOK, resp)
}

// JSONEvent represents a event in management API
type JSONEvent struct {
	ID        int64        `json:"id"`
	Action    string       `json:"action"`
	Target    nulls.JSON   `json:"target"`
	Device    string       `json:"device"`
	IP        nulls.String `json:"ip"`
	Result    string       `json:"result"`
	CreatedAt time.Time    `json:"created_at"`
}

// NewJSONEvent converts a Event into JSONEvent.
func NewJSONEvent(event *Event) (JSONEvent, error) {
	j := JSONEvent{}
	err := copier.Copy(&j, event)
	// Set result field into string representation from type EventResult
	j.Result = event.Result.String()
	j.Device = event.UserAgent.String
	return j, errors.Wrap(err, errors.ErrorUnknown, "")
}
