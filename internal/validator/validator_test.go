package validator

import (
	"testing"

	"authcore.io/authcore/pkg/nulls"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestValidatePhone(t *testing.T) {

	type TestPhone struct {
		Phone nulls.String `validate:"phone"`
	}

	testPhonePassVector := [...]string{"+85221234567", "+886955626788", "+14153435465", "+863154312154", "+441506392361", "+12435468764"}
	testPhoneFailVector := [...]string{"85221234567", "+852123", "+1438", "++85221234567", "+886955626788545654632", "+11536YDBNQOA", "+86 3154 312 154", "+1(252)555-1114"}

	for _, testVector := range testPhonePassVector {
		testPhonePass := &TestPhone{
			Phone: nulls.String{
				String: testVector,
				Valid:  true,
			},
		}
		err := Validate.Struct(testPhonePass)
		assert.Nil(t, err)
	}

	for _, testVector := range testPhoneFailVector {
		testPhoneFail := &TestPhone{
			Phone: nulls.String{
				String: testVector,
				Valid:  true,
			},
		}
		err := Validate.Struct(testPhoneFail)
		assert.Error(t, err)
	}

}

func TestValidateEmail(t *testing.T) {

	type TestEmail struct {
		Email nulls.String `validate:"email"`
	}

	testEmailPass := &TestEmail{
		Email: nulls.String{
			String: "samuel@blocksq.com",
			Valid:  true,
		},
	}

	testEmailFail := &TestEmail{
		Email: nulls.String{
			String: "samuel@blocksq.com@blocksq.com",
			Valid:  true,
		},
	}

	err := Validate.Struct(testEmailPass)
	assert.Nil(t, err)

	err = Validate.Struct(testEmailFail)
	assert.Error(t, err)
}

func TestValidateByte(t *testing.T) {
	type TestByte struct {
		Byte []byte `validate:"byte=32"`
	}

	type TestByteString struct {
		Byte nulls.String `validate:"byte=32"`
	}

	testBytePass := TestByte{
		Byte: []byte("49XZGM0cqpNxl3nPj8TLmFXREivD9lsqmOfsmTrnU84"),
	}

	testByteFail := TestByte{
		Byte: []byte("49XZGM0cqpNxl3nPj8TLmFXREivD9lsqmOf"),
	}

	err := Validate.Struct(testBytePass)
	assert.Nil(t, err)

	err = Validate.Struct(testByteFail)
	assert.Error(t, err)

	testByteStringPass := &TestByteString{
		Byte: nulls.String{
			String: "49XZGM0cqpNxl3nPj8TLmFXREivD9lsqmOfsmTrnU84",
			Valid:  true,
		},
	}

	testByteStringFail := &TestByteString{
		Byte: nulls.String{
			String: "49XZGM0cqpNxl3nPj8TLmFXREivD9lsqmOf",
			Valid:  true,
		},
	}

	err = Validate.Struct(testByteStringPass)
	assert.Nil(t, err)

	err = Validate.Struct(testByteStringFail)
	assert.Error(t, err)
}

func TestRedirectURIValiator(t *testing.T) {
	viper.Set("applications.authcore-io.allowed_callback_urls", []string{"http://example.com/"})

	type TestRedirectURI struct {
		ClientID           string `validate:"client_id"`
		SuccessRedirectURL string `validate:"success_redirect_url"`
	}

	err := Validate.Struct(&TestRedirectURI{
		ClientID:           "authcore-io",
		SuccessRedirectURL: "http://example.com/",
	})
	assert.Nil(t, err)

	err = Validate.Struct(&TestRedirectURI{
		ClientID:           "authcore-io",
		SuccessRedirectURL: "http://evil.com/",
	})
	assert.Error(t, err)

	err = Validate.Struct(&TestRedirectURI{
		ClientID:           "an-invalid-client",
		SuccessRedirectURL: "http://example.com/",
	})
	assert.Error(t, err)
}
