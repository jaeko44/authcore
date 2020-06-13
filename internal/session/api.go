package session

import (
	"net/http"
	"strconv"
	"time"

	"authcore.io/authcore/internal/apiutil"
	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/internal/user"
	"authcore.io/authcore/pkg/nulls"
	"authcore.io/authcore/pkg/paging"
	"github.com/jinzhu/copier"
	"github.com/labstack/echo/v4"
)

// APIv2 returns a function that registers API 2.0 endpoints with an Echo instance.
func APIv2(store *Store) func(e *echo.Echo) {
	return func(e *echo.Echo) {
		h := &handler{store: store}

		g := e.Group("/api/v2")
		g.GET("/users/:id/sessions", h.ListUserSessions)
		//g.POST("/users/:id/sessions/create", h.CreateUserSession)
		g.GET("/sessions/:id", h.GetSession)
		g.DELETE("/sessions/:id", h.DeleteSession)

		g.GET("/users/current/sessions", h.ListCurrentUserSessions)
		g.DELETE("/users/current/sessions/:id", h.DeleteCurrentUserSession)
		g.GET("/sessions/current", h.GetCurrentSession)
		g.DELETE("/sessions/current", h.DeleteCurrentSession)
	}
}

type handler struct {
	store *Store
}

func (h *handler) ListUserSessions(c echo.Context) error {
	id := c.Param("id")
	r := UserSessionQuery{}
	if err := c.Bind(&r); err != nil {
		return err
	}
	pageOption := r.PageOptions()
	ctx := c.Request().Context()
	sessions, page, err := h.store.FindAllSessionsByUser(ctx, pageOption, id)
	if err != nil {
		return err
	}
	jsonSessions := make([]JSONSession, len(*sessions))
	for i, s := range *sessions {
		jsonSessions[i], err = NewJSONSession(&s)
		if err != nil {
			return err
		}
	}
	resp := apiutil.NewListPagination(jsonSessions, page)
	return c.JSON(http.StatusOK, resp)
}

func (h *handler) ListCurrentUserSessions(c echo.Context) error {
	me, ok := user.FromContext(c)
	if !ok {
		return errors.New(errors.ErrorUnauthenticated, "")
	}

	ctx := c.Request().Context()
	pageOptions := paging.PageOptions{}
	sessions, page, err := h.store.FindAllSessionsByUser(ctx, pageOptions, me.PublicID())
	if err != nil {
		return err
	}
	jsonSessions := make([]JSONSession, len(*sessions))
	for i, s := range *sessions {
		jsonSessions[i], err = NewJSONSession(&s)
		if err != nil {
			return err
		}
	}
	resp := apiutil.NewListPagination(jsonSessions, page)
	return c.JSON(http.StatusOK, resp)
}

func (h *handler) GetSession(c echo.Context) error {
	id := c.Param("id")
	ctx := c.Request().Context()
	session, err := h.store.FindSessionByPublicID(ctx, id)
	if err != nil {
		return err
	}

	jsonSession, err := NewJSONSession(session)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, jsonSession)
}

func (h *handler) GetCurrentSession(c echo.Context) error {
	session, ok := FromContext(c)
	if !ok {
		return errors.New(errors.ErrorUnauthenticated, "")
	}

	jsonSession, err := NewJSONSession(session)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, jsonSession)
}

func (h *handler) DeleteSession(c echo.Context) error {
	s := c.Param("id")
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return errors.Wrap(err, errors.ErrorInvalidArgument, "")
	}
	ctx := c.Request().Context()
	_, err = h.store.InvalidateSessionByID(ctx, id)
	if err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *handler) DeleteCurrentSession(c echo.Context) error {
	session, ok := FromContext(c)
	if !ok {
		return errors.New(errors.ErrorUnauthenticated, "")
	}

	ctx := c.Request().Context()
	_, err := h.store.InvalidateSessionByID(ctx, session.ID)
	if err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *handler) DeleteCurrentUserSession(c echo.Context) error {
	me, ok := user.FromContext(c)
	if !ok {
		return errors.New(errors.ErrorUnauthenticated, "")
	}

	s := c.Param("id")
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return errors.Wrap(err, errors.ErrorInvalidArgument, "")
	}

	ctx := c.Request().Context()
	session, err := h.store.FindSessionByInternalID(ctx, id)
	if err != nil {
		return err
	}
	if session.UserID != me.ID {
		// Return 404 as if the session ID doesn't exist to avoid guessing the ID.
		return errors.New(errors.ErrorNotFound, "")
	}
	_, err = h.store.InvalidateSessionByID(ctx, id)
	if err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

// UserSessionQuery represent a query to user sessions
type UserSessionQuery struct {
	PageToken string `query:"page_token"`
	Limit     uint   `query:"limit" validate:"omitempty,gte=0,lte=1000"`
}

// PageOptions returns a PageOptions for the query.
func (q *UserSessionQuery) PageOptions() paging.PageOptions {
	limit := q.Limit
	if limit <= 0 || limit > 1000 {
		limit = 50
	}
	return paging.PageOptions{
		UniqueColumn:   "id",
		CountFoundRows: true,
		SortDirection:  paging.Desc,
		SortColumn:     "last_seen_at",
		Limit:          limit,
		PageToken:      paging.PageToken(q.PageToken),
	}
}

// JSONSession represents a session in management API.
type JSONSession struct {
	ID         int64        `json:"id"`
	ClientID   nulls.String `json:"client_id"`
	DeviceID   nulls.String `json:"device_id"`
	LastSeenAt time.Time    `json:"last_seen_at"`
	LastSeenIP string       `json:"last_seen_ip"`
	UserAgent  string       `json:"user_agent"`
}

// NewJSONSession converts a Session into JSONSession.
func NewJSONSession(s *Session) (JSONSession, error) {
	j := JSONSession{}
	err := copier.Copy(&j, s)
	return j, errors.Wrap(err, errors.ErrorUnknown, "")
}
