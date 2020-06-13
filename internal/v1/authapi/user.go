package authapi

import (
	"context"
	"sync"

	"authcore.io/authcore/internal/db"
	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/internal/user"
	"authcore.io/authcore/internal/user/registration"
	"authcore.io/authcore/internal/webhook"
	"authcore.io/authcore/pkg/api/authapi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/golang/protobuf/ptypes/timestamp"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// CreateUser function register the account into the system.
func (s *Service) CreateUser(ctx context.Context, in *authapi.CreateUserRequest) (*authapi.CreateUserResponse, error) {
	u := &user.User{
		Username:    db.NullableString(in.Username),
		DisplayNameOld: in.DisplayName,
		Email:       db.NullableString(in.Email),
		Phone:       db.NullableString(in.Phone),
		// Fixed value for language field. Support for deprecated API
		Language: db.NullableString(viper.GetStringSlice("available_languages")[0]),
	}
	clientID := in.ClientId
	session, err := registration.RegisterUser(ctx, s.UserStore, s.SessionStore, s.EmailService, s.SMSService, u, clientID, in.SendVerification, true)
	if err != nil {
		return nil, err
	}

	pbUser, err := MarshalUser(u)
	if err != nil {
		return nil, err
	}

	return &authapi.CreateUserResponse{
		User:         pbUser,
		RefreshToken: session.RefreshToken,
	}, nil
}

// GetCurrentUser returns the user information of the user authenticated by access token.
func (s *Service) GetCurrentUser(ctx context.Context, in *empty.Empty) (*authapi.User, error) {
	currentUser, ok := user.CurrentUserFromContext(ctx)
	if !ok || currentUser == nil {
		return nil, errors.New(errors.ErrorUnauthenticated, "")
	}

	pbUser, err := MarshalUser(currentUser)
	if err != nil {
		return nil, err
	}

	return pbUser, nil
}

// UpdateCurrentUser updates the current user profile
func (s *Service) UpdateCurrentUser(ctx context.Context, in *authapi.UpdateCurrentUserRequest) (*authapi.User, error) {
	currentUser, ok := user.CurrentUserFromContext(ctx)
	if !ok || currentUser == nil {
		return nil, errors.New(errors.ErrorUnauthenticated, "")
	}

	// Only allow to update display name, email and phone for the user
	if in.User.DisplayName != "" {
		currentUser.DisplayNameOld = in.User.DisplayName
	}

	err := s.UserStore.UpdateUser(ctx, currentUser)
	if err != nil {
		return nil, err
	}

	// FIXME: this code is unintentionally blocking
	var wg sync.WaitGroup
	wg.Add(1)
	go func(*user.User) {
		webhookResponse, err := webhook.MarshalUpdateUserResponse(currentUser)
		if err != nil {
			log.Error("cannot marshal webhook update user response")
			return
		}
		webhook.CallExternalWebhook(webhook.UpdateUserEvent, webhookResponse)
		wg.Done()
	}(currentUser)

	pbUser, err := MarshalUser(currentUser)
	if err != nil {
		return nil, err
	}

	wg.Wait()
	return pbUser, nil
}

// StartVerifyPrimaryContact verifies the primary contact
func (s *Service) StartVerifyPrimaryContact(ctx context.Context, in *authapi.StartVerifyPrimaryContactRequest) (*authapi.StartVerifyPrimaryContactResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "removed from v1.0")
}

// CompleteVerifyPrimaryContact verifies the primary contact
func (s *Service) CompleteVerifyPrimaryContact(ctx context.Context, in *authapi.CompleteVerifyPrimaryContactRequest) (*authapi.CompleteVerifyPrimaryContactResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "removed from v1.0")
}

// CheckHandleAvailability ...
func (s *Service) CheckHandleAvailability(ctx context.Context, in *authapi.CheckHandleAvailabilityRequest) (*authapi.CheckHandleAvailabilityResponse, error) {
	return &authapi.CheckHandleAvailabilityResponse{}, nil
}

// MarshalUser marshals *user.User into Protobuf message
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
		CreatedAt: &timestamp.Timestamp{
			Seconds: in.CreatedAt.Unix(),
			Nanos:   int32(in.CreatedAt.Nanosecond()),
		},
		UpdatedAt: &timestamp.Timestamp{
			Seconds: in.UpdatedAt.Unix(),
			Nanos:   int32(in.UpdatedAt.Nanosecond()),
		},
		PasswordAuthentication: in.IsPasswordAuthenticationEnabled(),
		Language:               in.RealLanguage(),
	}, nil
}
