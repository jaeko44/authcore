package validator

import (
	"database/sql/driver"
	"encoding/base64"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"authcore.io/authcore/internal/clientapp"
	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/pkg/httputil"
	"authcore.io/authcore/pkg/nulls"

	"github.com/go-playground/validator/v10"
	"github.com/nyaruka/phonenumbers"
	log "github.com/sirupsen/logrus"
)

// Validate is a Validate with additional validators
var Validate *validator.Validate

// Validator is a struct that implements echo.Validator
var Validator = echoValidator{}

type echoValidator struct{}

// Validate validates a given struct.
func (echoValidator) Validate(i interface{}) error {
	return errors.WithValidateError(Validate.Struct(i))
}

func init() {
	Validate = validator.New()
	Validate.RegisterValidation("phone", PhoneValidator)
	Validate.RegisterValidation("language", languageValidator)
	Validate.RegisterValidation("byte", byteLengthFromStringValidator)
	Validate.RegisterValidation("oauth_user_id", oauthUserIDValidator)
	Validate.RegisterValidation("challenge_set", challengeSetValidator)
	Validate.RegisterValidation("client_id", clientIDValidator)
	Validate.RegisterValidation("success_redirect_url", redirectURIValidator)
	Validate.RegisterCustomTypeFunc(valuerCustomTypeFunc, nulls.String{})
}

// PhoneValidator checks the phone format
func PhoneValidator(fl validator.FieldLevel) bool {
	phoneRegexString := "^\\+[0-9]+$"
	phoneRegex := regexp.MustCompile(phoneRegexString)
	if !phoneRegex.MatchString(fl.Field().String()) {
		return false
	}
	phoneNumber, err := phonenumbers.Parse(fl.Field().String(), "ZZ") // "ZZ" for unknown region
	if err != nil {
		return false
	}
	return phonenumbers.IsPossibleNumber(phoneNumber)
}

// languageValidator checks the language is valid or not
func languageValidator(fl validator.FieldLevel) bool {
	// Check the language format according BCP47
	// It supports for simple language subtag(e.g. 'en') and subtag plus Script subtag(e.g. 'en-US')
	if fl.Field().String() == "" {
		return true
	}
	languageRegex := regexp.MustCompile("^[a-z]+(-[a-zA-Z0-9]+|)$")
	if !languageRegex.MatchString(fl.Field().String()) {
		return false
	}
	return true
}

// valuerCustomTypeFunc implements validator.CustomTypeFunc for type Valuer in sql/driver
// Valuer is the generic function to retrun generic type Value using in sql/driver.
func valuerCustomTypeFunc(field reflect.Value) interface{} {
	if valuer, ok := field.Interface().(driver.Valuer); ok {
		val, err := valuer.Value()
		if err == nil {
			return val
		}
		// handle the error how you want
	}
	return nil
}

// byteLengthFromStringValidator checks the length of bytes from string
func byteLengthFromStringValidator(fl validator.FieldLevel) bool {
	field := fl.Field()
	param := fl.Param()

	switch field.Kind() {
	case reflect.String:
		p := parseIntOrPanic(param)
		result, _ := base64.RawURLEncoding.DecodeString(field.String())
		return int64(len(result)) == p
	case reflect.Slice:
		p := parseIntOrPanic(param)
		bytesResult := field.Bytes()
		result, _ := base64.RawURLEncoding.DecodeString(string(bytesResult))
		return int64(len(result)) == p
	}

	log.WithFields(log.Fields{
		"type": field.Interface(),
	}).Panic("Bad field type")
	return false
}

// oauthUserIDValidator checks if the user id >= 1 or the state is for OAuth.
func oauthUserIDValidator(fl validator.FieldLevel) bool {
	userID := fl.Field().Int()
	challenges := reflect.Indirect(fl.Parent()).FieldByName("Challenges")
	isOAuth := challenges.Len() == 1 && challenges.Index(0).String() == "OAUTH"
	return (isOAuth && userID == 0) || (!isOAuth && userID > 0)
}

// challengeSetValidate validates if "OAUTH" (the only challenge that user id is not required) is not used along other challenges
func challengeSetValidator(fl validator.FieldLevel) bool {
	challenges := fl.Field()
	challengeCount := challenges.Len()
	if challengeCount > 1 {
		for i := 0; i < challengeCount; i++ {
			if challenges.Index(i).String() == "OAUTH" {
				return false
			}
		}
	}
	return true
}

func clientIDValidator(fl validator.FieldLevel) bool {
	clientID := fl.Field().String()
	if strings.Contains(clientID, ".") {
		return false
	}
	_, err := clientapp.GetByClientID(clientID)
	if err != nil {
		return false
	}
	return true
}

func redirectURIValidator(fl validator.FieldLevel) bool {
	uri := fl.Field().String()
	if uri == "" {
		return true
	}

	clientID := reflect.Indirect(fl.Parent()).FieldByName("ClientID").String()
	if strings.Contains(clientID, ".") {
		return false
	}
	clientApp, err := clientapp.GetByClientID(clientID)
	if err != nil {
		return false
	}

	acceptURIPrefixes := clientApp.AllowedCallbackURLs
	normalizedURI, err := httputil.NormalizeURI(uri)
	if err != nil {
		return false
	}
	for _, acceptURIPrefix := range acceptURIPrefixes {
		if strings.HasPrefix(normalizedURI, acceptURIPrefix) {
			return true
		}
	}
	return false
}

// parseIntOrPanic returns the parameter as a int64
// or panics if it can't convert
func parseIntOrPanic(param string) int64 {

	i, err := strconv.ParseInt(param, 0, 64)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Panic("cannot parse the input to int64")
	}

	return i
}
