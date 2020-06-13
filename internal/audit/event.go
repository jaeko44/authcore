package audit

import (
	"strconv"
	"time"

	"authcore.io/authcore/pkg/nulls"
	"authcore.io/authcore/pkg/paging"
)

// EventResult is a type enumerating the result return from the API endpoint.
type EventResult int32

// Enumerates the EventResult
const (
	EventResultUnknown EventResult = 0
	EventResultFail    EventResult = 1
	EventResultSuccess EventResult = 2
)

func (result EventResult) String() string {
	switch result {
	case EventResultFail:
		return "fail"
	case EventResultSuccess:
		return "success"
	default:
		return "unknown"
	}
}

// SystemActor represents the actor of the log is system. It is represented as NULL in MySQL.
var SystemActor = nulls.Int64{}

// Event represents an audit log information of an action.
type Event struct {
	ID           int64        `db:"id"`
	ActorID      nulls.Int64  `db:"actor_id" fieldtag:"insert"`
	ActorDisplay nulls.String `db:"actor_display" fieldtag:"insert"`
	Action       string       `db:"action" validate:"required" fieldtag:"insert"`
	Target       nulls.JSON   `db:"target" fieldtag:"insert"`
	Result       EventResult  `db:"result" fieldtag:"insert"`
	IP           nulls.String `db:"ip" validate:"omitempty,ip" fieldtag:"insert"`
	UserAgent    nulls.String `db:"user_agent" fieldtag:"insert"`
	CreatedAt    time.Time    `db:"created_at"`
}

// PublicID returns a ID string that is suitable to be used by clients.
func (event *Event) PublicID() string {
	return strconv.FormatInt(event.ID, 10)
}

// Actor is an interface that represents the actor in an audit event.
type Actor interface {
	ActorID() int64
	DisplayName() string
}

// EventsQuery is a query for selecting audit logs Events.
type EventsQuery struct {
	PageToken string `query:"page_token"`
	Limit     uint   `query:"limit" validate:"omitempty,gte=0,lte=1000"`
	ActorID   string `query:"user_id"`
}

// PageOptions returns a PageOptions for the query.
func (q *EventsQuery) PageOptions() paging.PageOptions {
	limit := q.Limit
	if limit <= 0 || limit > 1000 {
		limit = 50
	}
	return paging.PageOptions{
		SortColumn:     "created_at",
		SortDirection:  paging.Desc,
		UniqueColumn:   "id",
		CountFoundRows: true,
		Limit:          limit,
		PageToken:      paging.PageToken(q.PageToken),
	}
}
