package authapi

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"authcore.io/authcore/internal/user"
	"authcore.io/authcore/pkg/api/authapi"
	"authcore.io/authcore/pkg/cryptoutil"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Creates a TOTP second factor
func TestCreateTOTPSecondFactor(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	u, err := srv.UserStore.UserByID(context.Background(), 1)
	assert.NoError(t, err)
	ctx := user.NewContextWithCurrentUser(context.Background(), u)

	// 1. List the second factors
	req := &authapi.ListSecondFactorsRequest{}
	res, err := srv.ListSecondFactors(ctx, req)
	if !assert.NoError(t, err) {
		return
	}
	assert.Len(t, res.SecondFactors, 0)

	// 2. Create a TOTP second factor
	secret := "THISISATOTPSECRETXXXXXXXXXXXXXXX"
	req2 := &authapi.CreateSecondFactorRequest{
		Info: &authapi.CreateSecondFactorRequest_TotpInfo{
			TotpInfo: &authapi.TOTPInfo{
				Secret:     secret,
				Identifier: "jPhone",
			},
		},
		Answer: cryptoutil.GetTOTPPin(secret, time.Now()),
	}
	_, err = srv.CreateSecondFactor(ctx, req2)
	if !assert.NoError(t, err) {
		return
	}

	// 3. List the second factors again
	req3 := &authapi.ListSecondFactorsRequest{}
	res3, err := srv.ListSecondFactors(ctx, req3)
	if !assert.NoError(t, err) {
		return
	}
	assert.Len(t, res3.SecondFactors, 1)
	secondFactor := res3.SecondFactors[0]
	assert.Equal(t, "jPhone", secondFactor.Content.Identifier)
}

// Fails to create a TOTP second factor given wrong PIN
func TestCreateTOTPSecondFactorWithWrongPIN(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	u, err := srv.UserStore.UserByID(context.Background(), 1)
	assert.NoError(t, err)
	ctx := user.NewContextWithCurrentUser(context.Background(), u)

	// 1. Create a TOTP second factor with the wrong PIN
	secret := "THISISATOTPSECRETXXXXXXXXXXXXXXX"
	req := &authapi.CreateSecondFactorRequest{
		Info: &authapi.CreateSecondFactorRequest_TotpInfo{
			TotpInfo: &authapi.TOTPInfo{
				Secret:     secret,
				Identifier: "jPhone",
			},
		},
		Answer: "000000",
	}
	_, err = srv.CreateSecondFactor(ctx, req)
	assert.Error(t, err)
	status, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, status.Code())
}

// Creates a SMS second factor
func TestCreateSMSSecondFactor(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()
	viper.Set("contact_verification_expiry_for_phone", "5s")

	u, err := srv.UserStore.UserByID(context.Background(), 1)
	assert.NoError(t, err)
	ctx := user.NewContextWithCurrentUser(context.Background(), u)

	// 1. List the second factors
	req := &authapi.ListSecondFactorsRequest{}
	res, err := srv.ListSecondFactors(ctx, req)
	assert.NoError(t, err)
	assert.Len(t, res.SecondFactors, 0)

	// 2. Request a SMS code for creating SMS second factor
	phoneNumber := "+85290000000"
	req2 := &authapi.StartCreateSecondFactorRequest{
		Info: &authapi.StartCreateSecondFactorRequest_SmsInfo{
			SmsInfo: &authapi.SMSInfo{
				PhoneNumber: phoneNumber,
			},
		},
	}
	_, err = srv.StartCreateSecondFactor(ctx, req2)
	assert.NoError(t, err)

	// 3. Create a SMS second factor
	closedLoopCodePartialKey := srv.UserStore.GetClosedLoopCodePartialKeyBySecondFactorValue(phoneNumber)
	closedLoopCode, err := srv.UserStore.FindClosedLoopCodeByKey(context.Background(), closedLoopCodePartialKey)
	assert.NoError(t, err)

	req3 := &authapi.CreateSecondFactorRequest{
		Info: &authapi.CreateSecondFactorRequest_SmsInfo{
			SmsInfo: &authapi.SMSInfo{
				PhoneNumber: phoneNumber,
			},
		},
		Answer: closedLoopCode.Code,
	}
	_, err = srv.CreateSecondFactor(ctx, req3)
	assert.NoError(t, err)

	// 4. List the second factors again
	req4 := &authapi.ListSecondFactorsRequest{}
	res4, err := srv.ListSecondFactors(ctx, req4)
	assert.NoError(t, err)
	assert.Len(t, res4.SecondFactors, 1)
	secondFactor := res4.SecondFactors[0]
	assert.Equal(t, phoneNumber, secondFactor.Content.PhoneNumber)
}

// Fails to create a SMS second factor given wrong PIN
func TestCreateSMSSecondFactorWithWrongPIN(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	u, err := srv.UserStore.UserByID(context.Background(), 1)
	assert.NoError(t, err)
	ctx := user.NewContextWithCurrentUser(context.Background(), u)

	// 1. Request a SMS code for creating SMS second factor
	phoneNumber := "+85290000000"
	req := &authapi.StartCreateSecondFactorRequest{
		Info: &authapi.StartCreateSecondFactorRequest_SmsInfo{
			SmsInfo: &authapi.SMSInfo{
				PhoneNumber: phoneNumber,
			},
		},
	}
	_, err = srv.StartCreateSecondFactor(ctx, req)
	assert.NoError(t, err)

	// 2. Create a SMS second factor
	req2 := &authapi.CreateSecondFactorRequest{
		Info: &authapi.CreateSecondFactorRequest_SmsInfo{
			SmsInfo: &authapi.SMSInfo{
				PhoneNumber: phoneNumber,
			},
		},
		Answer: "000000",
	}
	_, err = srv.CreateSecondFactor(ctx, req2)
	assert.Error(t, err)
}

// Creates a backup code second factor
func TestCreateBackupCodeSecondFactor(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	u, err := srv.UserStore.UserByID(context.Background(), 1)
	assert.NoError(t, err)
	ctx := user.NewContextWithCurrentUser(context.Background(), u)

	// 1. List the second factors
	req := &authapi.ListSecondFactorsRequest{}
	res, err := srv.ListSecondFactors(ctx, req)
	assert.NoError(t, err)
	assert.Len(t, res.SecondFactors, 0)

	// 2. Create a backup code second factor
	req2 := &authapi.CreateSecondFactorRequest{
		Info: &authapi.CreateSecondFactorRequest_BackupCodeInfo{
			BackupCodeInfo: &authapi.BackupCodeInfo{},
		},
		Answer: "",
	}
	_, err = srv.CreateSecondFactor(ctx, req2)
	assert.NoError(t, err)

	// 3. List the second factors again
	req3 := &authapi.ListSecondFactorsRequest{}
	res3, err := srv.ListSecondFactors(ctx, req3)
	assert.NoError(t, err)
	assert.Len(t, res3.SecondFactors, 1)
}

// Creates a backup code second factor twice
// The first should pass and the second should fail, as there should be at most one backup code second factor.
func TestCreateBackupCodeSecondFactorTwice(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	u, err := srv.UserStore.UserByID(context.Background(), 1)
	assert.NoError(t, err)
	ctx := user.NewContextWithCurrentUser(context.Background(), u)

	// 1. Create a backup code second factor
	req := &authapi.CreateSecondFactorRequest{
		Info: &authapi.CreateSecondFactorRequest_BackupCodeInfo{
			BackupCodeInfo: &authapi.BackupCodeInfo{},
		},
		Answer: "",
	}
	_, err = srv.CreateSecondFactor(ctx, req)
	assert.NoError(t, err)

	// 2. Create a backup code second factor
	req2 := &authapi.CreateSecondFactorRequest{
		Info: &authapi.CreateSecondFactorRequest_BackupCodeInfo{
			BackupCodeInfo: &authapi.BackupCodeInfo{},
		},
		Answer: "",
	}
	_, err = srv.CreateSecondFactor(ctx, req2)
	assert.Error(t, err)
}

// Lists the second factors of the current user
func TestListSecondFactors(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	u, err := srv.UserStore.UserByID(context.Background(), 4)
	assert.NoError(t, err)
	ctx := user.NewContextWithCurrentUser(context.Background(), u)

	// 1. List the second factors of TwoFactor
	req := &authapi.ListSecondFactorsRequest{}
	res, err := srv.ListSecondFactors(ctx, req)
	assert.NoError(t, err)
	assert.Len(t, res.SecondFactors, 3)
	jRes, err := json.Marshal(res)
	assert.NoError(t, err)
	assert.Contains(t, string(jRes), "TwoFactor's jPhone")                  // contains the identifier
	assert.NotContains(t, string(jRes), "THISISAWEAKTOTPSECRETFORTESTSXY2") // does not contain the secret
}

// Lists the second factors of the current user by type
func TestListSecondFactorsByType(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	u, err := srv.UserStore.UserByID(context.Background(), 4)
	assert.NoError(t, err)
	ctx := user.NewContextWithCurrentUser(context.Background(), u)

	// 1. List the SMS second factors of TwoFactor
	req := &authapi.ListSecondFactorsRequest{
		Type: "sms",
	}
	res, err := srv.ListSecondFactors(ctx, req)
	assert.NoError(t, err)
	assert.Len(t, res.SecondFactors, 1)
	secondFactor := res.SecondFactors[0]
	assert.Equal(t, "+85298765432", secondFactor.Content.PhoneNumber)
}

// Deletes a second factor of the current user
func TestDeleteSecondFactor(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	u, err := srv.UserStore.UserByID(context.Background(), 4)
	assert.NoError(t, err)
	ctx := user.NewContextWithCurrentUser(context.Background(), u)

	// 1. List the second factors
	req := &authapi.ListSecondFactorsRequest{}
	res, err := srv.ListSecondFactors(ctx, req)
	assert.NoError(t, err)
	assert.Len(t, res.SecondFactors, 3)

	// 2. Delete a second factor
	req2 := &authapi.DeleteSecondFactorRequest{
		Id: "3",
	}
	_, err = srv.DeleteSecondFactor(ctx, req2)
	assert.NoError(t, err)

	// 3. List the second factors again
	req3 := &authapi.ListSecondFactorsRequest{}
	res3, err := srv.ListSecondFactors(ctx, req3)
	assert.NoError(t, err)
	assert.Len(t, res3.SecondFactors, 2)
}

// Fails to delete a second factors for an another user
func TestDeleteSecondFactorForAnotherUser(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	u, err := srv.UserStore.UserByID(context.Background(), 1)
	assert.NoError(t, err)
	ctx := user.NewContextWithCurrentUser(context.Background(), u)

	// 1. Delete a second factor of TwoFactor, yet the current user is Bob
	req := &authapi.DeleteSecondFactorRequest{
		Id: "3",
	}
	_, err = srv.DeleteSecondFactor(ctx, req)
	assert.Error(t, err)
	status, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, status.Code())
}
