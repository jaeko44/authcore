package authentication

import (
	"sync"

	validator "gopkg.in/go-playground/validator.v9"
)

var validatorOnce sync.Once
var validate *validator.Validate

// Validate return the instance of validator for database.
func Validate() *validator.Validate {
	validatorOnce.Do(func() {
		validate = validator.New()
	})
	return validate
}
