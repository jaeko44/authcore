package audit

import (
	"context"
	"strconv"

	"authcore.io/authcore/internal/apiutil"
	"authcore.io/authcore/internal/db"
	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/pkg/log"
	"authcore.io/authcore/pkg/nulls"
	"authcore.io/authcore/pkg/paging"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/metadata"
)

// Store provides store audit events.
type Store struct {
	db *db.DB
}

// NewStore returns a new Store instance.
func NewStore(db *db.DB) *Store {
	db.AddModel(new(Event), "audit_logs", "id")
	return &Store{
		db: db,
	}
}

// InsertEvent is a low-level API to create a new event entry. It requires all audit log information to be provided.
func (s *Store) InsertEvent(ctx context.Context, event *Event) error {
	log.GetLogger(ctx).WithFields(logrus.Fields{
		"actor_id":   event.ActorID.Int64,
		"actor":      event.ActorDisplay.String,
		"action":     event.Action,
		"target":     event.Target.Struct,
		"result":     event.Result.String(),
		"ip":         event.IP.String,
		"user_agent": event.UserAgent.String,
	}).Info("audit event")
	return s.db.InsertModel(ctx, event)
}

// LogEvent inserts an audit event. If actor is nil, this method will attempt to get the
// authenticated user from context. Target must be an JSON serializable struct.
func (s *Store) LogEvent(c echo.Context, actor Actor, action string, target interface{}, result EventResult) {
	var actorID nulls.Int64
	var actorDisplay nulls.String
	if actor == nil {
		actor, _ = c.Get("user").(Actor)
	}
	if actor != nil {
		actorID = nulls.NewInt64(actor.ActorID())
		actorDisplay = nulls.NewString(actor.DisplayName())
	}

	targetJSON := nulls.JSON{}
	if target != nil {
		targetJSON = nulls.NewJSON(target)
	}
	ip := c.RealIP()
	ua := apiutil.FormatUserAgent(c.Request().UserAgent())
	event := &Event{
		ActorID:      actorID,
		ActorDisplay: actorDisplay,
		Action:       action,
		Target:       targetJSON,
		Result:       result,
		IP:           nulls.NewString(ip),
		UserAgent:    nulls.NewString(ua),
	}

	err := s.InsertEvent(c.Request().Context(), event)
	if err != nil {
		log.GetLogger(c.Request().Context()).Errorf("error writing audit log event: %v", err)
	}
}

// CreateEvent (deprecated) is a high-level API which create audit log in a generalized way. Input
// context should include IP and user agent as these information are fetched from context.
// GeneralTarget will be used if input target implements GetGeneralTarget method. Otherwise it will
// marshal to JSON value.
func (s *Store) CreateEvent(ctx context.Context, actor nulls.Int64, action string, _ nulls.Int64, target interface{}, result EventResult) error {
	event := &Event{
		ActorID: actor,
		Action:  action,
		Result:  result,
	}
	event.IP = nulls.String{}
	device := nulls.String{}
	// FIXME: The method to get IP address and agent is deprecated
	// Get the IP address and agent from context
	fromMD, ok := metadata.FromIncomingContext(ctx)
	if ok {
		ipAddress := fromMD.Get("ip-address")
		if len(ipAddress) > 0 {
			event.IP = nulls.NewString(ipAddress[0])
		}
		userAgent := fromMD.Get("grpcgateway-user-agent")
		device = nulls.NewString(apiutil.FormatUserAgent(userAgent[0]))
	}
	event.UserAgent = device
	event.Target = nulls.NewJSON(target)

	return s.InsertEvent(ctx, event)
}

// AllEventsWithPageOptions lookups all Events in the database.
func (s *Store) AllEventsWithPageOptions(ctx context.Context, pageOptions paging.PageOptions) (*[]Event, *paging.Page, error) {
	sb, err := s.db.ModelMap.SelectAllBuilder(new(Event))
	if err != nil {
		return nil, nil, err
	}

	sq, sa := sb.Build()
	events := []Event{}
	page, err := paging.SelectContext(ctx, s.db, pageOptions, &events, sq, sa...)
	if err != nil {
		return nil, nil, err
	}
	return &events, page, nil
}

// AllEventsWithQuery lookup Events with EventsQuery.
func (s *Store) AllEventsWithQuery(ctx context.Context, query EventsQuery) (*[]Event, *paging.Page, error) {
	sb, err := s.db.ModelMap.SelectAllBuilder(new(Event))
	if err != nil {
		return nil, nil, err
	}

	if query.ActorID != "" {
		actorID, err := strconv.ParseInt(query.ActorID, 10, 64)
		if err != nil {
			return nil, nil, errors.New(errors.ErrorInvalidArgument, "invalid actor_id in query")
		}
		sb.Where(sb.Like("actor_id", actorID))
	}
	sq, sa := sb.Build()
	pageOptions := query.PageOptions()
	events := []Event{}
	page, err := paging.SelectContext(ctx, s.db, pageOptions, &events, sq, sa...)
	if err != nil {
		return nil, nil, err
	}

	return &events, page, nil
}
