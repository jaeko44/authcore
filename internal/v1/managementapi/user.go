package managementapi

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"authcore.io/authcore/internal/v1/authentication"
	"authcore.io/authcore/internal/clientapp"
	"authcore.io/authcore/internal/db"
	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/internal/user"
	"authcore.io/authcore/internal/user/registration"
	"authcore.io/authcore/internal/webhook"
	"authcore.io/authcore/pkg/api/authapi"
	"authcore.io/authcore/pkg/api/managementapi"
	"authcore.io/authcore/pkg/nulls"
	"authcore.io/authcore/pkg/paging"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// ListUsers returns a list of users according to id in descending order.
// The number of returned users can be set by PageSize and specific page can be set by PageToken.
// The user accessing have to be authenticated by access token.
func (s *Service) ListUsers(ctx context.Context, in *managementapi.ListUsersRequest) (*managementapi.ListUsersResponse, error) {
	// Valid sort keys for filter options
	validSortKeys := []string{
		"is_locked",
		"created_at",
		"last_seen_at",
		"email",
	}
	// Valid query keys for filter options
	validQueryKeys := []string{
		"all",
		"email",
		"phone",
		"username",
		"display_name",
	}
	currentUser, ok := user.CurrentUserFromContext(ctx)
	if !ok || currentUser == nil {
		return nil, errors.New(errors.ErrorUnauthenticated, "")
	}

	err := s.authorize(ctx, ListUsersPermission)
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

	sortDirection := "desc"
	if in.Ascending {
		sortDirection = "asc"
	}

	if in.SortKey == "" {
		in.SortKey = "last_seen_at"
	}

	sortKeyValidity := false
	for _, validSortKey := range validSortKeys {
		sortKeyValidity = strings.EqualFold(validSortKey, in.SortKey)
		if sortKeyValidity {
			break
		}
	}

	if !sortKeyValidity {
		return nil, errors.New(errors.ErrorInvalidArgument, "SortKey is invalid")
	}

	if in.QueryKey == "" {
		in.QueryKey = "all"
	}

	queryKeyValidity := false
	for _, validQueryKey := range validQueryKeys {
		queryKeyValidity = strings.EqualFold(validQueryKey, in.QueryKey)
		if queryKeyValidity {
			break
		}
	}

	if !queryKeyValidity {
		return nil, errors.New(errors.ErrorInvalidArgument, "QueryKey is invalid")
	}

	usersQuery := user.UsersQuery{
		Limit:     uint(in.PageSize),
		PageToken: in.PageToken,
		SortBy:    fmt.Sprintf("%s %s", in.SortKey, sortDirection),
	}
	if in.QueryKey == "email" {
		usersQuery.Email = in.QueryValue
	}
	if in.QueryKey == "phone" {
		usersQuery.Phone = in.QueryValue
	}
	if in.QueryKey == "display_name" {
		usersQuery.Name = in.QueryValue
	}
	if in.QueryKey == "username" {
		usersQuery.Username = in.QueryValue
	}

	var users *[]user.User
	var page *paging.Page
	if in.UserHandle != "" {
		u, err := s.UserStore.UserByHandle(ctx, in.UserHandle)
		if errors.IsKind(err, errors.ErrorNotFound) {
			users = &[]user.User{}
		} else if u != nil {
			users = &[]user.User{*u}
		}
		page = &paging.Page{FoundRows: 1}
	} else {
		users, page, err = s.UserStore.AllUsersWithQuery(ctx, usersQuery)
	}
	if err != nil {
		return nil, err
	}

	var pbUsers []*authapi.User
	for _, user := range *users {
		pbUser, err := MarshalUser(&user)
		if err != nil {
			return nil, err
		}
		pbUsers = append(pbUsers, pbUser)
	}

	return &managementapi.ListUsersResponse{
		Users:             pbUsers,
		NextPageToken:     string(page.NextPageToken),
		PreviousPageToken: string(page.PreviousPageToken),
		TotalSize:         int32(page.FoundRows),
	}, nil
}

// CreateUser function register the account into the system.
func (s *Service) CreateUser(ctx context.Context, in *managementapi.CreateUserRequest) (*managementapi.CreateUserResponse, error) {
	currentUser, ok := user.CurrentUserFromContext(ctx)
	if !ok || currentUser == nil {
		return nil, errors.New(errors.ErrorUnauthenticated, "")
	}

	err := s.authorize(ctx, CreateUserPermission)
	if err != nil {
		return nil, err
	}

	u := &user.User{
		Username:       db.NullableString(in.Username),
		DisplayNameOld: in.DisplayName,
		Email:          db.NullableString(in.Email),
		Phone:          db.NullableString(in.Phone),
		// Fixed value for language field. Support for deprecated API
		Language: db.NullableString(viper.GetStringSlice("available_languages")[0]),
	}

	// check if oauth factors are used, and there are at most one user id in each of the oauth services
	for i := 0; i < len(in.OauthFactors); i++ {
		for j := i + 1; j < len(in.OauthFactors); j++ {
			if in.OauthFactors[i].Service == in.OauthFactors[j].Service {
				return nil, errors.New(errors.ErrorInvalidArgument, "there should be at most one account per oauth service")
			}
		}
	}
	for _, oauthFactor := range in.OauthFactors {
		if oauthFactor.OauthUserId == "" {
			return nil, errors.New(errors.ErrorInvalidArgument, "oauth id could not be empty")
		}
		_, err = s.UserStore.FindOAuthFactorByOAuthIdentity(ctx, user.OAuthService(oauthFactor.Service), oauthFactor.OauthUserId)
		if err != nil && !errors.IsKind(err, errors.ErrorNotFound) {
			return nil, err
		} else if err == nil {
			return nil, errors.New(errors.ErrorAlreadyExists, "")
		}
	}
	clientID := clientapp.AdminPortalClientID
	session, err := registration.RegisterUser(ctx, s.UserStore, s.SessionStore, s.EmailService, s.SMSService, u, clientID, false, true)
	if err != nil {
		return nil, err
	}

	// create oauth factors
	for _, oauthFactor := range in.OauthFactors {
		_, err = s.UserStore.CreateOAuthFactor(ctx, u.ID, user.OAuthService(oauthFactor.Service), oauthFactor.OauthUserId, nulls.NewJSON(oauthFactor.Metadata))
		if err != nil {
			return nil, err
		}
	}

	pbUser, err := MarshalUser(u)
	if err != nil {
		return nil, err
	}

	return &managementapi.CreateUserResponse{
		User:         pbUser,
		RefreshToken: session.RefreshToken,
	}, nil
}

// GetUser returns information of a user
func (s *Service) GetUser(ctx context.Context, in *managementapi.GetUserRequest) (*authapi.User, error) {
	currentUser, ok := user.CurrentUserFromContext(ctx)
	if !ok || currentUser == nil {
		return nil, errors.New(errors.ErrorUnauthenticated, "")
	}

	err := s.authorize(ctx, GetUserPermission)
	if err != nil {
		return nil, err
	}

	user, err := s.UserStore.UserByPublicID(ctx, in.UserId)
	if err != nil {
		return nil, err
	}

	pbUser, err := MarshalUser(user)
	if err != nil {
		return nil, err
	}

	return pbUser, nil
}

// UpdateUser updates user information.
func (s *Service) UpdateUser(ctx context.Context, in *managementapi.UpdateUserRequest) (*authapi.User, error) {
	currentUser, ok := user.CurrentUserFromContext(ctx)
	if !ok || currentUser == nil {
		return nil, errors.New(errors.ErrorUnauthenticated, "")
	}

	u, err := s.UserStore.UserByPublicID(ctx, in.UserId)
	if err != nil {
		return nil, err
	}

	if in.User == nil {
		return nil, errors.New(errors.ErrorInvalidArgument, "")
	}

	switch in.Type {
	case managementapi.UpdateUserRequest_PROFILE:
		err := s.authorize(ctx, UpdateUserPermission)
		if err != nil {
			return nil, err
		}
		u.Username = db.NullableString(in.User.Username)
		u.Language = db.NullableString(in.User.Language)
		u.DisplayNameOld = in.User.DisplayName
	case managementapi.UpdateUserRequest_LOCK:
		err := s.authorize(ctx, LockUserPermission)
		if err != nil {
			return nil, err
		}
		var lockExpiredAt time.Time
		if in.User.LockExpiredAt == nil {
			lockExpiredAt, err = time.Parse(time.RFC3339, "0001-01-01T00:00:00Z")
		} else {
			lockExpiredAt, err = ptypes.Timestamp(in.User.LockExpiredAt)
		}
		if err != nil {
			return nil, errors.Wrap(err, errors.ErrorUnknown, "")
		}
		u.IsLocked = in.User.Locked
		u.LockExpiredAt = db.NullableTime(lockExpiredAt)
		u.LockDescription = db.NullableString(in.User.LockDescription)
	}

	err = s.UserStore.UpdateUser(ctx, u)
	if err != nil {
		return nil, err
	}

	// FIXME: this code is unintentionally blocking
	var wg sync.WaitGroup
	wg.Add(1)
	go func(*user.User) {
		webhookResponse, err := webhook.MarshalUpdateUserResponse(u)
		if err != nil {
			log.Error("cannot marshal webhook update user response")
			return
		}
		webhook.CallExternalWebhook(webhook.UpdateUserEvent, webhookResponse)
		wg.Done()
	}(u)

	pbUser, err := MarshalUser(u)
	if err != nil {
		return nil, err
	}

	wg.Wait()
	return pbUser, nil
}

// CreateFirstAdminUser creates the first admin user for a deployment. If a user is already registered, this method
// returns an error. This method is intented to be called by CLI .
func (s *Service) CreateFirstAdminUser(ctx context.Context, u *user.User, password string) (*user.User, error) {
	users, _, err := s.UserStore.AllUsersWithQuery(ctx, user.UsersQuery{})
	if err != nil {
		return nil, err
	}
	if len(*users) > 0 {
		return nil, errors.New(errors.ErrorAlreadyExists, "")
	}

	if password == "" {
		return nil, errors.New(errors.ErrorInvalidArgument, "")
	}

	adminRole, err := s.UserStore.FindRoleByName(ctx, "authcore.admin")
	if err != nil {
		return nil, err
	}

	clientID := clientapp.AdminPortalClientID
	_, err = registration.RegisterUser(ctx, s.UserStore, s.SessionStore, s.EmailService, s.SMSService, u, clientID, false, false)
	if err != nil {
		return nil, err
	}

	clientIdentity, serverIdentity := authentication.GetIdentity()
	spake2, err := authentication.NewSPAKE2Plus()
	if err != nil {
		return nil, err
	}

	salt, verifierW0, verifierL, err := authentication.GenerateSPAKE2Verifier([]byte(password), clientIdentity, serverIdentity, spake2)
	if err != nil {
		return nil, err
	}

	err = u.SetPasswordVerifier(salt, verifierW0, verifierL)
	if err != nil {
		return nil, errors.New(errors.ErrorInvalidArgument, "")
	}

	err = s.UserStore.UpdateUser(ctx, u)
	if err != nil {
		return nil, err
	}

	err = s.UserStore.AssignRole(ctx, &user.RoleUser{
		RoleID: adminRole.ID,
		UserID: u.ID,
	})
	if err != nil {
		return nil, err
	}

	// Refresh user
	u, err = s.UserStore.UserByID(ctx, u.ID)
	if err != nil {
		return nil, err
	}

	return u, nil
}

// MarshalUser marshals *User into Protobuf message
func MarshalUser(in *user.User) (*authapi.User, error) {
	return &authapi.User{
		Id:           in.PublicID(),
		Username:     in.Username.String,
		PrimaryEmail: in.Email.String,
		PrimaryEmailVerified: &timestamp.Timestamp{
			Seconds: in.EmailVerifiedAt.Time.Unix(),
			Nanos:   int32(in.EmailVerifiedAt.Time.Nanosecond()),
		},
		PrimaryPhone: in.Phone.String,
		PrimaryPhoneVerified: &timestamp.Timestamp{
			Seconds: in.PhoneVerifiedAt.Time.Unix(),
			Nanos:   int32(in.PhoneVerifiedAt.Time.Nanosecond()),
		},
		RecoveryEmail: in.RecoveryEmail.String,
		RecoveryEmailVerified: &timestamp.Timestamp{
			Seconds: in.RecoveryEmailVerifiedAt.Time.Unix(),
			Nanos:   int32(in.RecoveryEmailVerifiedAt.Time.Nanosecond()),
		},
		DisplayName: in.DisplayNameOld,
		LastSeenAt: &timestamp.Timestamp{
			Seconds: in.LastSeenAt.Unix(),
			Nanos:   int32(in.LastSeenAt.Nanosecond()),
		},
		CreatedAt: &timestamp.Timestamp{
			Seconds: in.CreatedAt.Unix(),
			Nanos:   int32(in.CreatedAt.Nanosecond()),
		},
		UpdatedAt: &timestamp.Timestamp{
			Seconds: in.UpdatedAt.Unix(),
			Nanos:   int32(in.UpdatedAt.Nanosecond()),
		},
		Locked:                 in.IsCurrentlyLocked(),
		PasswordAuthentication: in.IsPasswordAuthenticationEnabled(),
		Language:               in.RealLanguage(),
	}, nil
}
