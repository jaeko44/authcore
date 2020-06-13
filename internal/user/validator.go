package user

import (
	v "authcore.io/authcore/internal/validator"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

func init() {
	v.Validate.RegisterStructValidation(userStructLevelValidation, &User{})
}

// userStructLevelValidation contains customized validation logic for username, email and phone fields which fits with different user handle scenarios.
func userStructLevelValidation(sl validator.StructLevel) {
	user := sl.Current().Interface().(User)

	requireUserEmailOrPhone := viper.GetBool("require_user_email_or_phone")
	requireUserPhone := viper.GetBool("require_user_phone")
	requireUserEmail := viper.GetBool("require_user_email")
	requireUserUsername := viper.GetBool("require_user_username")

	// Set default user email requirement if setting is not found
	if !(requireUserEmailOrPhone || requireUserPhone || requireUserEmail || requireUserUsername) {
		requireUserEmail = true
	}

	phoneErr := sl.Validator().Var(user.Phone, "required")
	emailErr := sl.Validator().Var(user.Email, "required")
	usernameErr := sl.Validator().Var(user.Username, "required")

	if requireUserEmailOrPhone {
		if phoneErr != nil && emailErr != nil {
			sl.ReportError(user.Phone, "Phone", "phone", "phone", "contact is missing")
			sl.ReportError(user.Email, "Email", "email", "email", "contact is missing")
		}
	}

	if requireUserEmail {
		if emailErr != nil {
			sl.ReportError(user.Email, "Email", "email", "email", "contact is missing")
		}
	}

	if requireUserPhone {
		if phoneErr != nil {
			sl.ReportError(user.Phone, "Phone", "phone", "phone", "contact is missing")
		}
	}

	if requireUserUsername {
		if usernameErr != nil {
			sl.ReportError(user.Username, "Username", "required", "required", "")
		}
	}
}
