package verifier

import (
	"testing"
	"time"

	"authcore.io/authcore/internal/config"
	"authcore.io/authcore/internal/db"
	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/internal/sms"
	"authcore.io/authcore/internal/template"
	"authcore.io/authcore/internal/testutil"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func factoryForTest() (*Factory, func()) {
	config.InitDefaults()
	viper.Set("secret_key_base", "855edf399835e9c9deb61877c1a76bf14eed7c35a167e10ff1b7d43db4363268")
	viper.Set("base_path", "../../..")
	config.InitConfig()
	testutil.FixturesSetUp()
	d := db.NewDBFromConfig()
	redis := testutil.RedisForTest()
	templateStore := template.NewStore(d)
	smsService := sms.NewService(templateStore)
	f := NewFactory()
	f.Register(SMSOTP, SMSOTPVerifierFactory(smsService, redis))

	return f, func() {
		d.Close()
		viper.Reset()
		redis.FlushAll()
	}
}

func TestSMSOTPVerifier(t *testing.T) {
	f, teardown := factoryForTest()
	defer teardown()

	data := `{
		"method": "sms_otp",
		"phone_number": "+85212345678"
	}`
	verifier, err := f.Unmarshal([]byte(data))
	assert.NoError(t, err)
	_, ok := verifier.(SMSOTPVerifier)
	assert.True(t, ok)
	assert.Equal(t, "sms_otp", verifier.Method())
	assert.False(t, verifier.IsPrimary())
	assert.False(t, verifier.SkipMFA())
	assert.Empty(t, verifier.Salt())

	vs, challenge, err := verifier.Request(nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, vs)
	assert.Empty(t, challenge)

	// Incorrect OTP
	assert.NoError(t, err)
	ok, _ = verifier.Verify(vs, []byte("123456"))
	assert.False(t, ok)
	ok, _ = verifier.Verify(vs, nil)
	assert.False(t, ok)

	// Incorrect VerifierState
	ok, _ = verifier.Verify([]byte("xxx"), []byte("123456"))
	assert.False(t, ok)
	ok, _ = verifier.Verify(nil, []byte("123456"))
	assert.False(t, ok)
	ok, _ = verifier.Verify(nil, nil)
	assert.False(t, ok)

	// Correct OTP
	cs, err := codeStateFromState(vs)
	assert.NoError(t, err)
	ok, vs2 := verifier.Verify(vs, []byte(cs.Code))
	assert.True(t, ok)
	assert.Nil(t, vs2)
}

func TestSMSOTPVerifierTooManyRequests(t *testing.T) {
	f, teardown := factoryForTest()
	defer teardown()

	data := `{
		"method": "sms_otp",
		"phone_number": "+85212345678"
	}`
	verifier, err := f.Unmarshal([]byte(data))
	assert.NoError(t, err)
	_, ok := verifier.(SMSOTPVerifier)
	assert.True(t, ok)

	_, _, err = verifier.Request(nil)
	assert.NoError(t, err)

	_, _, err = verifier.Request(nil)
	assert.Error(t, err)
	assert.True(t, errors.IsKind(err, errors.ErrorResourceExhausted))
}

func TestSMSOTPVerifierOTPExpire(t *testing.T) {
	f, teardown := factoryForTest()
	defer teardown()
	viper.Set("sms_code_expiry", "100ms")

	data := `{
		"method": "sms_otp",
		"phone_number": "+85212345678"
	}`
	verifier, err := f.Unmarshal([]byte(data))
	assert.NoError(t, err)
	_, ok := verifier.(SMSOTPVerifier)
	assert.True(t, ok)

	vs, _, err := verifier.Request(nil)
	assert.NoError(t, err)

	time.Sleep(500 * time.Millisecond)

	cs, err := codeStateFromState(vs)
	assert.NoError(t, err)
	ok, vs2 := verifier.Verify(vs, []byte(cs.Code))
	assert.False(t, ok)
	assert.Nil(t, vs2)
}
