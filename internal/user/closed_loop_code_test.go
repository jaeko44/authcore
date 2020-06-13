package user

//FIXME: this file should be moved to another package

import (
	"context"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestCreateClosedLoopCode(t *testing.T) {
	store, teardown := storeForTest()
	defer teardown()

	viper.Set("closed_loop_code_length", "5")
	viper.Set("closed_loop_max_attempts", "123")

	expiry, err := time.ParseDuration("1h")
	assert.Nil(t, err)

	closedLoopCode, err := store.CreateClosedLoopCode(context.Background(), "contact:3", expiry)
	if assert.Nil(t, err) {
		assert.NotNil(t, closedLoopCode)
		assert.Equal(t, "contact:3", closedLoopCode.Key)
		assert.Equal(t, expiry, closedLoopCode.CodeExpireAt.Sub(closedLoopCode.CodeSentAt))
		assert.Len(t, closedLoopCode.Code, 5)
		assert.Equal(t, int64(123), closedLoopCode.RemainingAttempts)
	}
}

func TestFindClosedLoopCodeByKey(t *testing.T) {
	store, teardown := storeForTest()
	defer teardown()

	expiry, err := time.ParseDuration("10m")
	assert.Nil(t, err)

	closedLoopCode, err := store.CreateClosedLoopCode(context.Background(), "contact:3", expiry)
	assert.Nil(t, err)

	closedLoopCode2, err := store.FindClosedLoopCodeByKey(context.Background(), "contact:3")
	assert.Nil(t, err)

	assert.Equal(t, closedLoopCode.Key, closedLoopCode2.Key)
	assert.Equal(t, closedLoopCode.CodeSentAt.Unix(), closedLoopCode2.CodeSentAt.Unix())
	assert.Equal(t, closedLoopCode.CodeExpireAt.Unix(), closedLoopCode2.CodeExpireAt.Unix())
	assert.Equal(t, closedLoopCode.Code, closedLoopCode2.Code)
	assert.Equal(t, closedLoopCode.Token, closedLoopCode2.Token)
	assert.Equal(t, closedLoopCode.RemainingAttempts, closedLoopCode2.RemainingAttempts)
}

func TestFindClosedLoopCodeByToken(t *testing.T) {
	store, teardown := storeForTest()
	defer teardown()

	expiry, err := time.ParseDuration("10m")
	assert.Nil(t, err)

	closedLoopCode, err := store.CreateClosedLoopCode(context.Background(), "contact:3", expiry)
	assert.Nil(t, err)

	closedLoopCode2, err := store.FindClosedLoopCodeByToken(context.Background(), closedLoopCode.Token)
	assert.Nil(t, err)

	assert.Equal(t, closedLoopCode.Key, closedLoopCode2.Key)
	assert.Equal(t, closedLoopCode.CodeSentAt.Unix(), closedLoopCode2.CodeSentAt.Unix())
	assert.Equal(t, closedLoopCode.CodeExpireAt.Unix(), closedLoopCode2.CodeExpireAt.Unix())
	assert.Equal(t, closedLoopCode.Code, closedLoopCode2.Code)
	assert.Equal(t, closedLoopCode.Token, closedLoopCode2.Token)
	assert.Equal(t, closedLoopCode.RemainingAttempts, closedLoopCode2.RemainingAttempts)
}

func TestGetLastClosedLoopCodeSentAt(t *testing.T) {
	store, teardown := storeForTest()
	defer teardown()

	expiry, err := time.ParseDuration("10m")
	assert.Nil(t, err)

	closedLoopCode, err := store.CreateClosedLoopCode(context.Background(), "contact:3", expiry)
	assert.Nil(t, err)

	lastCodeSentAt, err := store.GetClosedLoopCodeLastSentAt(context.Background(), "contact:3")
	assert.Nil(t, err)

	assert.Equal(t, closedLoopCode.CodeSentAt.Unix(), lastCodeSentAt.Unix())
}

func TestBurnClosedLoopCodeWithCode(t *testing.T) {
	store, teardown := storeForTest()
	defer teardown()

	viper.Set("closed_loop_max_attempts", "5")

	expiry, err := time.ParseDuration("10m")
	assert.Nil(t, err)

	closedLoopCode, err := store.CreateClosedLoopCode(context.Background(), "contact:3", expiry)
	assert.Nil(t, err)

	// The code is burned
	_, err = store.BurnClosedLoopCodeByCode(context.Background(), "contact:3", closedLoopCode.Code)
	assert.Nil(t, err)

	// The code cannot be reused
	_, err = store.BurnClosedLoopCodeByCode(context.Background(), "contact:3", closedLoopCode.Code)
	assert.NotNil(t, err)
}

func TestBurnClosedLoopCodeWithToken(t *testing.T) {
	store, teardown := storeForTest()
	defer teardown()

	viper.Set("closed_loop_max_attempts", "5")

	expiry, err := time.ParseDuration("10m")
	assert.Nil(t, err)

	closedLoopCode, err := store.CreateClosedLoopCode(context.Background(), "contact:3", expiry)
	assert.Nil(t, err)

	// The code is burned
	_, err = store.BurnClosedLoopCodeByToken(context.Background(), closedLoopCode.Token)
	assert.Nil(t, err)

	// The code cannot be reused
	_, err = store.BurnClosedLoopCodeByToken(context.Background(), closedLoopCode.Token)
	assert.NotNil(t, err)
}

func TestBurnClosedLoopCodeWithTooMuchIncorrectAttempts(t *testing.T) {
	store, teardown := storeForTest()
	defer teardown()

	viper.Set("closed_loop_code_length", "5")
	viper.Set("closed_loop_max_attempts", "1")

	expiry, err := time.ParseDuration("10m")
	assert.Nil(t, err)

	closedLoopCode, err := store.CreateClosedLoopCode(context.Background(), "contact:3", expiry)
	assert.Nil(t, err)

	for attempts := 1; attempts <= 10; attempts++ {
		// Invalid input
		_, err = store.BurnClosedLoopCodeByCode(context.Background(), "contact:3", "00000")
		assert.Error(t, err)
		// Cannot burn even the correct code is given, as there are too many incorrect attempts
		_, err = store.BurnClosedLoopCodeByCode(context.Background(), "contact:3", closedLoopCode.Code)
		assert.Error(t, err)
	}
}

func TestDecrementClosedLoopCodeRemainingAttempts(t *testing.T) {
	store, teardown := storeForTest()
	defer teardown()

	viper.Set("closed_loop_code_length", "5")
	viper.Set("closed_loop_max_attempts", "100")

	expiry, err := time.ParseDuration("10m")
	assert.Nil(t, err)

	closedLoopCode, err := store.CreateClosedLoopCode(context.Background(), "contact:3", expiry)
	assert.Nil(t, err)

	_, err = store.BurnClosedLoopCodeByCode(context.Background(), "contact:3", "00000")
	assert.Error(t, err)

	closedLoopCode2, err := store.FindClosedLoopCodeByKey(context.Background(), "contact:3")
	assert.Nil(t, err)
	assert.Equal(t, closedLoopCode.RemainingAttempts-1, closedLoopCode2.RemainingAttempts)
}

func TestClosedLoopCodeExpiry(t *testing.T) {
	store, teardown := storeForTest()
	defer teardown()

	expiry, err := time.ParseDuration("10m")
	assert.Nil(t, err)

	closedLoopCode, err := store.CreateClosedLoopCode(context.Background(), "contact:3", expiry)
	assert.Nil(t, err)

	assert.Equal(t, expiry, closedLoopCode.Expiry())
}

func TestGetCloseLoopCodePartialKeyByContactID(t *testing.T) {
	store, teardown := storeForTest()
	defer teardown()

	closeLoopCodePartialKey := store.GetClosedLoopCodePartialKeyByContactID(1337)
	assert.Equal(t, "contact_id:1337", closeLoopCodePartialKey)
}

func TestGetCloseLoopCodePartialKeyBySecondFactorID(t *testing.T) {
	store, teardown := storeForTest()
	defer teardown()

	closeLoopCodePartialKey := store.GetClosedLoopCodePartialKeyBySecondFactorID(1337)
	assert.Equal(t, "second_factor_id:1337", closeLoopCodePartialKey)
}

func TestGetCloseLoopCodePartialKeyBySecondFactorValue(t *testing.T) {
	store, teardown := storeForTest()
	defer teardown()

	closeLoopCodePartialKey := store.GetClosedLoopCodePartialKeyBySecondFactorValue("+85298765432")
	assert.Equal(t, "second_factor_value:+85298765432", closeLoopCodePartialKey)
}

func TestGetClosedLoopCodeContactID(t *testing.T) {
	store, teardown := storeForTest()
	defer teardown()

	expiry, err := time.ParseDuration("10m")
	assert.Nil(t, err)

	closedLoopCode, err := store.CreateClosedLoopCode(context.Background(), "contact_id:3", expiry)
	assert.Nil(t, err)

	contactID, err := closedLoopCode.GetContactID()
	assert.Nil(t, err)
	assert.Equal(t, int64(3), contactID)
}
