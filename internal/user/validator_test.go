package user

import (
	"testing"

	"authcore.io/authcore/internal/validator"
	"authcore.io/authcore/pkg/nulls"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestValidateUser(t *testing.T) {
	var err error
	// language
	viper.SetDefault("available_languages", []string{
		"en",
		"zh-HK",
	})

	viper.SetDefault("require_user_email", true)
	viper.SetDefault("require_user_phone", true)
	viper.SetDefault("require_user_username", true)

	user := &User{
		Username:     nulls.NewString("bob"),
		Email: nulls.NewString("bob@example.com"),
		Phone: nulls.NewString("+85221234567"),
		DisplayNameOld:  "Bob",
		Language: nulls.NewString("en"),
	}

	err = validator.Validate.Struct(user)
	assert.Nil(t, err)

	userNoEmail := &User{
		Username:     nulls.NewString("bob"),
		Email: nulls.NewString(""),
		Phone: nulls.NewString("+85221234567"),
		DisplayNameOld:  "Bob",
		Language: nulls.NewString("en"),
	}

	err = validator.Validate.Struct(userNoEmail)
	assert.Error(t, err)

	userNoPhone := &User{
		Username:     nulls.NewString("bob"),
		Email: nulls.NewString("bob@example.com"),
		Phone: nulls.NewString(""),
		DisplayNameOld:  "Bob",
		Language: nulls.NewString("en"),
	}

	err = validator.Validate.Struct(userNoPhone)
	assert.Error(t, err)

	userNoUsername := &User{
		Username:     nulls.NewString(""),
		Email: nulls.NewString("bob@example.com"),
		Phone: nulls.NewString("+85221234567"),
		DisplayNameOld:  "Bob",
		Language: nulls.NewString("en"),
	}

	err = validator.Validate.Struct(userNoUsername)
	assert.Error(t, err)

	viper.SetDefault("require_user_email_or_phone", false)

	viper.SetDefault("require_user_email", false)
	viper.SetDefault("require_user_phone", true)
	err = validator.Validate.Struct(userNoEmail)
	assert.Nil(t, err)
	err = validator.Validate.Struct(userNoPhone)
	assert.Error(t, err)

	viper.SetDefault("require_user_email", true)
	viper.SetDefault("require_user_phone", false)
	err = validator.Validate.Struct(userNoEmail)
	assert.Error(t, err)
	err = validator.Validate.Struct(userNoPhone)
	assert.Nil(t, err)

	viper.SetDefault("require_user_email", true)
	viper.SetDefault("require_user_phone", true)
	viper.SetDefault("require_user_username", false)
	err = validator.Validate.Struct(userNoEmail)
	assert.Error(t, err)
	err = validator.Validate.Struct(userNoPhone)
	assert.Error(t, err)
	err = validator.Validate.Struct(userNoUsername)
	assert.Nil(t, err)

	viper.SetDefault("require_user_email_or_phone", true)
	viper.SetDefault("require_user_email", false)
	viper.SetDefault("require_user_phone", false)
	viper.SetDefault("require_user_username", true)
	err = validator.Validate.Struct(userNoEmail)
	assert.Nil(t, err)
	err = validator.Validate.Struct(userNoPhone)
	assert.Nil(t, err)
}
