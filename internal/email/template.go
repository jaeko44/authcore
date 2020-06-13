package email

import (
	"html"

	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/internal/template"

	"github.com/spf13/viper"
	"golang.org/x/text/language"
)

type email struct {
	Subject string
	HTML    string
	Raw     string
}

type emailTemplate struct {
	Subject      template.String
	HTMLTemplate template.String
	RawTemplate  template.String
}

// Execute maps each TemplateString.Execute and return a email object
func (e emailTemplate) Execute(m map[string]string) email {
	Subject := e.Subject.Execute(map[string]string{})
	Raw := e.RawTemplate.Execute(m)
	for k, v := range m {
		m[k] = html.EscapeString(v)
	}
	HTML := e.HTMLTemplate.Execute(m)
	return email{
		Subject,
		HTML,
		Raw,
	}
}

func defaultLanguage() (language.Tag, error) {
	defaultLanguage := viper.GetString("default_language")
	tag, err := language.Parse(defaultLanguage)
	if err != nil {
		return language.Tag{}, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	return tag, nil
}
