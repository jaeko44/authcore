package managementapi

import (
	"context"
	"strconv"

	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/internal/clientapp"
	"authcore.io/authcore/internal/session"
	"authcore.io/authcore/internal/user"
	"authcore.io/authcore/pkg/api/authapi"
	"authcore.io/authcore/pkg/api/managementapi"
	"authcore.io/authcore/pkg/paging"

	"github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ListSessions lists active sessions of a specified user.
func (s *Service) ListSessions(ctx context.Context, in *managementapi.ListSessionsRequest) (*managementapi.ListSessionsResponse, error) {

	currentUser, ok := user.CurrentUserFromContext(ctx)
	if !ok || currentUser == nil {
		return nil, errors.New(errors.ErrorUnauthenticated, "")
	}

	err := s.authorize(ctx, ListSessionsPermission)
	if err != nil {
		return nil, err
	}

	sessions := &[]session.Session{}

	if in.PageSize <= 0 {
		// Default to be 10 entries for a page
		in.PageSize = 10
	}
	if in.PageSize > 1000 {
		return nil, status.Error(codes.InvalidArgument, "page size is too large")
	}

	sortDirection := paging.Desc
	if in.Ascending {
		sortDirection = paging.Asc
	}

	pageOptions := paging.PageOptions{
		Limit:          uint(in.PageSize),
		PageToken:      paging.PageToken(in.PageToken),
		SortDirection:  sortDirection,
		CountFoundRows: true,
	}

	sessions, page, err := s.SessionStore.FindAllSessionsByUser(ctx, pageOptions, in.UserId)
	if err != nil {
		return nil, err
	}

	var pbSessions []*authapi.Session
	for _, session := range *sessions {
		pbSession, err := MarshalSession(&session)
		if err != nil {
			return nil, err
		}
		pbSessions = append(pbSessions, pbSession)
	}

	return &managementapi.ListSessionsResponse{
		Sessions:          pbSessions,
		NextPageToken:     string(page.NextPageToken),
		PreviousPageToken: string(page.PreviousPageToken),
		TotalSize:         int32(page.FoundRows),
	}, nil
}

// CreateSession create an active session for user, return a new authenticated session for user
func (s *Service) CreateSession(ctx context.Context, in *managementapi.CreateSessionRequest) (*authapi.AccessToken, error) {
	currentUser, ok := user.CurrentUserFromContext(ctx)
	if !ok || currentUser == nil {
		return nil, errors.New(errors.ErrorUnauthenticated, "")
	}

	err := s.authorize(ctx, CreateSessionPermission)
	if err != nil {
		return nil, err
	}

	if in.UserId == "" || in.DeviceId == "" {
		return nil, errors.New(errors.ErrorInvalidArgument, "")
	}

	user, err := s.UserStore.UserByPublicID(ctx, in.UserId)
	if err != nil {
		return nil, err
	}

	deviceID, err := strconv.ParseInt(in.DeviceId, 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorInvalidArgument, "")
	}

	session, err := s.SessionStore.CreateSession(ctx, user.ID, deviceID, clientapp.AdminPortalClientID, "", false)
	if err != nil {
		return nil, err
	}

	bundle, err := s.SessionStore.GenerateAccessToken(ctx, session, true)
	if err != nil {
		return nil, err
	}

	return &authapi.AccessToken{
		AccessToken:  bundle.AccessToken,
		IdToken:      bundle.IDToken,
		RefreshToken: session.RefreshToken,
		TokenType:    authapi.AccessToken_BEARER,
		ExpiresIn:    bundle.ExpiresIn,
	}, nil
}

// DeleteSession invalidate an active session.
func (s *Service) DeleteSession(ctx context.Context, in *managementapi.DeleteSessionRequest) (*managementapi.DeleteSessionResponse, error) {
	currentUser, ok := user.CurrentUserFromContext(ctx)
	if !ok || currentUser == nil {
		return nil, errors.New(errors.ErrorUnauthenticated, "")
	}

	err := s.authorize(ctx, DeleteSessionPermission)
	if err != nil {
		return nil, err
	}

	session, err := s.SessionStore.FindSessionByPublicID(ctx, in.SessionId)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, errors.New(errors.ErrorPermissionDenied, "")
	}

	_, err = s.SessionStore.InvalidateSessionByID(ctx, session.ID)
	if err != nil {
		return nil, err
	}

	return &managementapi.DeleteSessionResponse{}, nil
}

// MarshalSession marshals *Session into Protobuf message
func MarshalSession(in *session.Session) (*authapi.Session, error) {
	return &authapi.Session{
		Id:     in.PublicID(),
		UserId: in.PublicUserID(),
		// TODO: Device association
		DeviceId:   "",
		LastSeenIp: in.LastSeenIP,
		LastSeenAt: &timestamp.Timestamp{
			Seconds: in.LastSeenAt.Unix(),
			Nanos:   int32(in.LastSeenAt.Nanosecond()),
		},
		LastSeenLocation: in.LastSeenLocation,
		UserAgent:        in.UserAgent,
		ExpiredAt: &timestamp.Timestamp{
			Seconds: in.ExpiredAt.Unix(),
			Nanos:   int32(in.ExpiredAt.Nanosecond()),
		},
		CreatedAt: &timestamp.Timestamp{
			Seconds: in.CreatedAt.Unix(),
			Nanos:   int32(in.CreatedAt.Nanosecond()),
		},
	}, nil
}
