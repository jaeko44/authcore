package email

import (
	"authcore.io/authcore/internal/errors"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// SendgridEmailProvider is a struct of Sendgrid email delegate.
type SendgridEmailProvider struct {
	client *sendgrid.Client
}

// NewSendgridEmailProvider creates a new SendgridEmailProvider with an api key.
func NewSendgridEmailProvider(apiKey string) (*SendgridEmailProvider, error) {
	client := sendgrid.NewSendClient(apiKey)
	return &SendgridEmailProvider{
		client: client,
	}, nil
}

// Send sends an email to a recipient with Sendgrid.
func (p SendgridEmailProvider) Send(from, to People, subject, rawBody, htmlBody string) error {
	sgFrom := mail.NewEmail(from.Name, from.Email)
	sgTo := mail.NewEmail(to.Name, to.Email)
	message := mail.NewSingleEmail(sgFrom, subject, sgTo, rawBody, htmlBody)

	_, err := p.client.Send(message)
	if err != nil {
		return errors.Wrap(err, errors.ErrorUnknown, "")
	}
	return nil
}
