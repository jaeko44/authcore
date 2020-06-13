package authapi

import (
	"context"

	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/internal/session"
	"authcore.io/authcore/internal/user"
	"authcore.io/authcore/pkg/api/authapi"
	"authcore.io/authcore/pkg/paging"

	"github.com/golang/protobuf/ptypes/timestamp"
)

// ListSessions lists all active (not invalidated, not expired) sessions for the current user.
func (s *Service) ListSessions(ctx context.Context, in *authapi.ListSessionsRequest) (*authapi.ListSessionsResponse, error) {
	currentUser, ok := user.CurrentUserFromContext(ctx)
	if !ok || currentUser == nil {
		return nil, errors.New(errors.ErrorUnauthenticated, "")
	}

	currentSession, ok := session.CurrentSessionFromContext(ctx)
	if !ok || currentSession == nil {
		return nil, errors.New(errors.ErrorUnauthenticated, "")
	}

	var err error
	sessions := &[]session.Session{}

	if in.PageSize <= 0 {
		// Default to be 100 entries for a page
		in.PageSize = 100
	}
	if in.PageSize > 1000 {
		return nil, errors.New(errors.ErrorInvalidArgument, "")
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

	sessions, page, err := s.SessionStore.FindAllSessionsByUser(ctx, pageOptions, currentUser.PublicID())
	if err != nil {
		return nil, err
	}

	var pbSessions []*authapi.Session
	for _, session := range *sessions {
		currentFlag := false
		if session.ID == currentSession.ID {
			currentFlag = true
		}
		pbSession, err := MarshalSession(&session, currentFlag)
		if err != nil {
			return nil, err
		}
		pbSessions = append(pbSessions, pbSession)
	}

	return &authapi.ListSessionsResponse{
		Sessions:          pbSessions,
		NextPageToken:     string(page.NextPageToken),
		PreviousPageToken: string(page.PreviousPageToken),
		TotalSize:         int32(page.FoundRows),
	}, nil
}

// DeleteSession deletes a session of the current user
func (s *Service) DeleteSession(ctx context.Context, in *authapi.DeleteSessionRequest) (*authapi.DeleteSessionResponse, error) {
	currentUser, ok := user.CurrentUserFromContext(ctx)
	if !ok || currentUser == nil {
		return nil, errors.New(errors.ErrorUnauthenticated, "")
	}

	session, err := s.SessionStore.FindSessionByPublicID(ctx, in.SessionId)
	if err != nil {
		return nil, err
	}
	if session == nil || session.UserID != currentUser.ID {
		return nil, errors.New(errors.ErrorUnauthenticated, "")
	}

	_, err = s.SessionStore.InvalidateSessionByID(ctx, session.ID)
	if err != nil {
		return nil, err
	}

	return &authapi.DeleteSessionResponse{}, nil
}

// DeleteCurrentSession deletes a current session of the user (signs out)
func (s *Service) DeleteCurrentSession(ctx context.Context, in *authapi.DeleteCurrentSessionRequest) (*authapi.DeleteCurrentSessionResponse, error) {
	currentUser, ok := user.CurrentUserFromContext(ctx)
	if !ok || currentUser == nil {
		return nil, errors.New(errors.ErrorUnauthenticated, "")
	}

	currentSession, ok := session.CurrentSessionFromContext(ctx)
	if !ok || currentSession == nil {
		return nil, errors.New(errors.ErrorUnauthenticated, "")
	}

	_, err := s.SessionStore.InvalidateSessionByID(ctx, currentSession.ID)
	if err != nil {
		return nil, err
	}

	return &authapi.DeleteCurrentSessionResponse{}, nil
}

// CreateMachineToken create a new machine token for the current user.
func (s *Service) CreateMachineToken(ctx context.Context, in *authapi.CreateMachineTokenRequest) (*authapi.CreateMachineTokenResponse, error) {
	currentUser, ok := user.CurrentUserFromContext(ctx)
	if !ok || currentUser == nil {
		return nil, errors.New(errors.ErrorUnauthenticated, "")
	}

	session, err := s.SessionStore.CreateMachineSession(ctx, currentUser.ID)
	if err != nil {
		return nil, err
	}

	return &authapi.CreateMachineTokenResponse{
		MachineToken: session.RefreshToken,
	}, nil
}

// MarshalSession marshals *Session into Protobuf message
func MarshalSession(in *session.Session, currentFlag bool) (*authapi.Session, error) {
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
		IsCurrent: currentFlag,
	}, nil
}
