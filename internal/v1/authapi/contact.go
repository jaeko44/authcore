package authapi

import (
	"context"

	"authcore.io/authcore/pkg/api/authapi"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreateContact creates a contact for the current user (TO-BE DEPRECATED)
func (s *Service) CreateContact(ctx context.Context, in *authapi.CreateContactRequest) (*authapi.Contact, error) {
	return nil, status.Errorf(codes.Unimplemented, "removed from v1.0")
}

// ListContacts lists the contacts for the current user (TO-BE DEPRECATED)
func (s *Service) ListContacts(ctx context.Context, in *authapi.ListContactsRequest) (*authapi.ListContactsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "removed from v1.0")
}

// DeleteContact deletes a contact of the current user
func (s *Service) DeleteContact(ctx context.Context, in *authapi.DeleteContactRequest) (*authapi.DeleteContactResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "removed from v1.0")
}

// UpdatePrimaryContact updates the specified contact as a primary contact of the current user. (TO-BE DEPRECATED)
func (s *Service) UpdatePrimaryContact(ctx context.Context, in *authapi.UpdatePrimaryContactRequest) (*authapi.Contact, error) {
	return nil, status.Errorf(codes.Unimplemented, "removed from v1.0")
}

// StartVerifyContact initiates the verifying process by sending an email / a SMS to the corresponding contact. (TO-BE DEPRECATED)
func (s *Service) StartVerifyContact(ctx context.Context, in *authapi.StartVerifyContactRequest) (*authapi.StartVerifyContactResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "removed from v1.0")
}

// CompleteVerifyContact checks if the verification code or the verification token is valid. (TO-BE DEPRECATED)
func (s *Service) CompleteVerifyContact(ctx context.Context, in *authapi.CompleteVerifyContactRequest) (*authapi.CompleteVerifyContactResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "removed from v1.0")
}
