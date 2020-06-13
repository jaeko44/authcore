package managementapi

import (
	"context"
	"strconv"

	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/internal/user"
	"authcore.io/authcore/pkg/api/authapi"
	"authcore.io/authcore/pkg/api/managementapi"

	"github.com/golang/protobuf/ptypes/timestamp"
	// log "github.com/sirupsen/logrus"
)

// ListSecondFactors lists the second factors for a given user
func (s *Service) ListSecondFactors(ctx context.Context, in *managementapi.ListSecondFactorsRequest) (*managementapi.ListSecondFactorsResponse, error) {
	currentUser, ok := user.CurrentUserFromContext(ctx)
	if !ok || currentUser == nil {
		return nil, errors.New(errors.ErrorUnauthenticated, "")
	}

	err := s.authorize(ctx, ListSecondFactorsPermission)
	if err != nil {
		return nil, err
	}

	userID, err := strconv.ParseInt(in.UserId, 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	secondFactors, err := s.UserStore.FindAllSecondFactorsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var pbSecondFactors []*authapi.SecondFactor
	for _, secondFactor := range *secondFactors {
		pbSecondFactor, err := MarshalSecondFactor(&secondFactor)
		if err != nil {
			return nil, err
		}
		pbSecondFactors = append(pbSecondFactors, pbSecondFactor)
	}

	return &managementapi.ListSecondFactorsResponse{
		SecondFactors: pbSecondFactors,
	}, nil
}

// MarshalSecondFactor marshals a *user.SecondFactor into Protobuf message
func MarshalSecondFactor(in *user.SecondFactor) (*authapi.SecondFactor, error) {
	return &authapi.SecondFactor{
		Id:     in.ID,
		UserId: in.UserID,
		Type:   authapi.SecondFactor_Type(in.Type),
		Content: &authapi.SecondFactor_Content{
			PhoneNumber: in.Content.PhoneNumber.String,
			Identifier:  in.Content.Identifier.String,
		},
		CreatedAt: &timestamp.Timestamp{
			Seconds: in.CreatedAt.Unix(),
			Nanos:   int32(in.CreatedAt.Nanosecond()),
		},
		LastUsedAt: &timestamp.Timestamp{
			Seconds: in.LastUsedAt.Unix(),
			Nanos:   int32(in.LastUsedAt.Nanosecond()),
		},
	}, nil
}
