package template

import (
	"regexp"
	"time"

	"github.com/spf13/viper"
)

// String is a abstraction of template string to do custom template functions.
type String struct {
	Template  string
	UpdatedAt time.Time
}

// Execute replace the placeholder in TemplateString by given map and returns the result.
func (t String) Execute(m map[string]string) string {
	ret := t.Template
	re := regexp.MustCompile(`{\w+}`)
	convert := func(s string) string {
		k := s[1 : len(s)-1]
		if val, ok := m[k]; ok {
			return val
		}
		return s
	}
	return re.ReplaceAllStringFunc(ret, convert)
}

// ToTemplateString convert string to TemplateString
func ToTemplateString(s string) String {
	return String{
		Template:  s,
		UpdatedAt: time.Unix(0, 0),
	}
}

// Type is a type enumerating the type for templates.
type Type int32

// Enumerates the TemplateType
const (
	TemplateSMS   Type = 0
	TemplateEMAIL Type = 1
)

// Template represent a template in DB.
type Template struct {
	ID             int64     `db:"id"`
	Name           string    `db:"name"`
	Language       string    `db:"language"`
	Template       string    `db:"template"`
	TemplateString String    ``
	UpdatedAt      time.Time `db:"updated_at"`
	CreatedAt      time.Time `db:"created_at"`
}

// Lists the available templates
var (
	EmailTemplates = []string{"VerificationMail", "ResetPasswordAuthenticationMail"}
	SMSTemplates   = []string{"AuthenticationSMS", "VerificationSMS", "ResetPasswordAuthenticationSMS"}
)

// GetAvailableLanguages retuns list of available languages in server
func GetAvailableLanguages() []string {
	return viper.GetStringSlice("available_languages")
}

// GetUpdatedAtFromStrings returns the latest "UpdatedAt" amongst the strings
func GetUpdatedAtFromStrings(strings ...String) *time.Time {
	updatedAt := time.Unix(0, 0)
	for _, str := range strings {
		if str.UpdatedAt.After(updatedAt) {
			updatedAt = str.UpdatedAt
		}
	}
	if updatedAt == time.Unix(0, 0) {
		return nil
	}
	return &updatedAt
}

// GetTemplateNamesFromType gets corresponding template name slice from Type
func GetTemplateNamesFromType(t Type) []string {
	switch t {
	case TemplateEMAIL:
		return EmailTemplates
	case TemplateSMS:
		return SMSTemplates
	}
	return nil
}
