package authapi

import (
	"context"

	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/pkg/api/authapi"
)

// CreateDeviceRegistration registers a device after satisfying anti-spam checks (e.g. proof-of-work).
func (s *Service) CreateDeviceRegistration(ctx context.Context, in *authapi.CreateDeviceRegistrationRequest) (*authapi.DeviceRegistration, error) {
	return nil, errors.New(errors.ErrorUnknown, "unimplemented")
}
