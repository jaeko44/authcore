package managementapi

import (
	"context"

	"authcore.io/authcore/pkg/api/authapi"
	"authcore.io/authcore/pkg/api/managementapi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreateContact create a contact for a user
func (s *Service) CreateContact(ctx context.Context, in *managementapi.CreateContactRequest) (*authapi.Contact, error) {
	return nil, status.Errorf(codes.Unimplemented, "removed from v1.0")
}

// ListContacts returns a list of contacts.
func (s *Service) ListContacts(ctx context.Context, in *managementapi.ListContactsRequest) (*managementapi.ListContactsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "removed from v1.0")
}

// DeleteContact delete a contact for a user
func (s *Service) DeleteContact(ctx context.Context, in *managementapi.DeleteContactRequest) (*managementapi.DeleteContactResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "removed from v1.0")
}

// StartVerifyContact initiates the verifying process by sending an email / a SMS to the corresponding contact.
func (s *Service) StartVerifyContact(ctx context.Context, in *managementapi.StartVerifyContactRequest) (*managementapi.StartVerifyContactResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "removed from v1.0")
}

// UpdatePrimaryContact updates primary contact of an user.
func (s *Service) UpdatePrimaryContact(ctx context.Context, in *managementapi.UpdatePrimaryContactRequest) (*authapi.Contact, error) {
	return nil, status.Errorf(codes.Unimplemented, "removed from v1.0")
}
