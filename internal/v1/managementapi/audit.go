package managementapi

import (
	"context"

	"authcore.io/authcore/internal/audit"
	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/internal/user"
	"authcore.io/authcore/pkg/api/authapi"
	"authcore.io/authcore/pkg/api/managementapi"
	"authcore.io/authcore/pkg/paging"

	"github.com/golang/protobuf/ptypes/timestamp"
)

// ListAuditLogs returns a list of audit logs according to id in descending order. It can be filtered by using UserId.
// The number of returned audit logs can be set by PageSize and specific page can be set by PageToken.
// The user accessing have to be authenticated by access token.
func (s *Service) ListAuditLogs(ctx context.Context, in *managementapi.ListAuditLogsRequest) (*managementapi.ListAuditLogsResponse, error) {
	var auditLogs *[]audit.Event
	var page *paging.Page
	var err error

	currentUser, ok := user.CurrentUserFromContext(ctx)
	if !ok || currentUser == nil {
		return nil, errors.New(errors.ErrorUnauthenticated, "")
	}

	err = s.authorize(ctx, ListAuditLogsPermission)
	if err != nil {
		return nil, err
	}

	if in.PageSize <= 0 {
		// Default to be 10 entries for a page
		in.PageSize = 10
	}

	if in.PageSize > 1000 {
		return nil, errors.New(errors.ErrorInvalidArgument, "")
	}

	eventsQuery := audit.EventsQuery {
		PageToken: in.PageToken,
		Limit: uint(in.PageSize),
	}

	if in.UserId != "" {
		eventsQuery.ActorID = in.UserId
	}

	auditLogs, page, err = s.AuditStore.AllEventsWithQuery(ctx, eventsQuery)
	if err != nil {
		return nil, err
	}

	var pbAuditLogs []*authapi.AuditLogEntity
	for _, auditLog := range *auditLogs {
		pbEvent, err := MarshalEvent(&auditLog)
		if err != nil {
			return nil, err
		}
		pbAuditLogs = append(pbAuditLogs, pbEvent)
	}
	return &managementapi.ListAuditLogsResponse{
		AuditLogs:         pbAuditLogs,
		NextPageToken:     string(page.NextPageToken),
		PreviousPageToken: string(page.PreviousPageToken),
		TotalSize:         int32(page.FoundRows),
	}, nil
}

// MarshalEvent marshals *Event into Protobuf message
func MarshalEvent(in *audit.Event) (*authapi.AuditLogEntity, error) {
	return &authapi.AuditLogEntity{
		Id:          in.PublicID(),
		Username:    in.ActorDisplay.String,
		Action:      in.Action,
		Target:      ToStruct(in.Target.Struct),
		Device:      in.UserAgent.String,
		Ip:          in.IP.String,
		Result:      authapi.AuditLogEntity_Result(in.Result),
		CreatedAt: &timestamp.Timestamp{
			Seconds: in.CreatedAt.Unix(),
			Nanos:   int32(in.CreatedAt.Nanosecond()),
		},
	}, nil
}
