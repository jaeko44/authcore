package authapi

import (
	"context"
	"testing"

	"authcore.io/authcore/pkg/api/authapi"

	"github.com/stretchr/testify/assert"
)

func TestValidateOAuthParameters(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	req := &authapi.ValidateOAuthParametersRequest{
		ClientId:     "authcore-io",
		ResponseType: "code",
		RedirectUri:  "http://0.0.0.0:8000/",
		State:        "YmccdxRKAs-c3OvQ6RKcKA",
		Scope:        "email",
	}
	_, err := srv.ValidateOAuthParameters(context.Background(), req)
	assert.NoError(t, err)

	req2 := &authapi.ValidateOAuthParametersRequest{
		ClientId:     "authcore-io",
		ResponseType: "code",
		RedirectUri:  "http://0.0.0.0:8001/",
		State:        "YmccdxRKAs-c3OvQ6RKcKA",
		Scope:        "email",
	}
	_, err = srv.ValidateOAuthParameters(context.Background(), req2)
	assert.Error(t, err)

	req3 := &authapi.ValidateOAuthParametersRequest{
		ClientId:            "authcore-io",
		ResponseType:        "code",
		RedirectUri:         "http://0.0.0.0:8000/",
		State:               "YmccdxRKAs-c3OvQ6RKcKA",
		Scope:               "email",
		CodeChallenge:       "bjQLnP-zepicpUTmu3gKLHiQHT-zNzh2hRGjBhevoB0",
		CodeChallengeMethod: "S256",
	}
	_, err = srv.ValidateOAuthParameters(context.Background(), req3)
	assert.NoError(t, err)

	req4 := &authapi.ValidateOAuthParametersRequest{
		ClientId:      "authcore-io",
		ResponseType:  "code",
		RedirectUri:   "http://0.0.0.0:8000/",
		State:         "YmccdxRKAs-c3OvQ6RKcKA",
		Scope:         "email",
		CodeChallenge: "bjQLnP-zepicpUTmu3gKLHiQHT-zNzh2hRGjBhevoB0",
	}
	_, err = srv.ValidateOAuthParameters(context.Background(), req4)
	assert.Error(t, err)

	// test without client id
	req5 := &authapi.ValidateOAuthParametersRequest{
		ResponseType: "code",
		RedirectUri:  "http://0.0.0.0:8000/",
		State:        "YmccdxRKAs-c3OvQ6RKcKA",
		Scope:        "email",
	}
	_, err = srv.ValidateOAuthParameters(context.Background(), req5)
	// should not error as fallback to default
	assert.NoError(t, err)
}
