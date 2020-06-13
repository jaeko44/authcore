package webhook

import (
	"encoding/json"

	"authcore.io/authcore/internal/user"
)

// UpdateUserResponseDataUser represents a user inside UpdateUserResponseData
type UpdateUserResponseDataUser struct {
	PublicID             string `json:"public_id"`
	PrimaryEmail         string `json:"primary_email"`
	PrimaryPhone         string `json:"primary_phone"`
	DisplayName          string `json:"display_name"`
	PrimaryEmailVerified bool   `json:"primary_email_verified"`
	PrimaryPhoneVerified bool   `json:"primary_phone_verified"`
}

// UpdateUserResponseData represents a data inside UpdateUserResponse
type UpdateUserResponseData struct {
	User UpdateUserResponseDataUser `json:"user"`
}

// UpdateUserResponse represents a response for UpdateUser event
type UpdateUserResponse struct {
	Data UpdateUserResponseData `json:"data"`
}

// MarshalUpdateUserResponse marshals into a update user response
func MarshalUpdateUserResponse(user *user.User) ([]byte, error) {
	updateUserResponse := UpdateUserResponse{
		Data: UpdateUserResponseData{
			User: UpdateUserResponseDataUser{
				PublicID:             user.PublicID(),
				PrimaryEmail:         user.Email.String,
				PrimaryPhone:         user.Phone.String,
				DisplayName:          user.DisplayName(),
				PrimaryEmailVerified: user.EmailVerifiedAt.Valid,
				PrimaryPhoneVerified: user.PhoneVerifiedAt.Valid,
			},
		},
	}
	jResponse, err := json.Marshal(updateUserResponse)
	if err != nil {
		return []byte{}, err
	}
	return jResponse, nil
}
