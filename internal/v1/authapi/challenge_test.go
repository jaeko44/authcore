package authapi

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestCreateProofOfWorkChallenge test response with decided pattern
func TestCreateProofOfWorkChallenge(t *testing.T) {
	assert := assert.New(t)
	srv, teardown := ServiceForTest()
	defer teardown()

	res, _ := srv.CreateProofOfWorkChallenge(context.Background(), &empty.Empty{})

	// Assert response token to be UUID format
	_, err := uuid.Parse(res.Token)
	assert.Nil(err)

	// Assert challenge is base64 encoded string
	_, err = base64.RawURLEncoding.DecodeString(res.Challenge)
	assert.Nil(err)
}
