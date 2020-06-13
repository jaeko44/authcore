package email

import (
	"net/mail"

	"authcore.io/authcore/internal/errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
)

// SESEmailProvider is a struct of AWS Simple Email Service (SES) email delegate.
type SESEmailProvider struct {
	service *ses.SES
}

// NewSESEmailProvider creates a new SESEmailProvider with region, access key id and secret access key.
func NewSESEmailProvider(region, accessKeyID, secretAccessKey string) (*SESEmailProvider, error) {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKeyID, secretAccessKey, ""),
	})
	if err != nil {
		return nil, err
	}
	service := ses.New(sess)

	return &SESEmailProvider{
		service: service,
	}, nil
}

// Send sends an email to a receipient with AWS SES.
func (p SESEmailProvider) Send(from, to People, subject, rawBody, htmlBody string) error {
	charset := "UTF-8"
	mail := &ses.SendEmailInput{
		Source: toAWSAddressString(from),
		Destination: &ses.Destination{
			CcAddresses: []*string{},
			ToAddresses: []*string{
				toAWSAddressString(to),
			},
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Html: &ses.Content{
					Charset: aws.String(charset),
					Data:    aws.String(htmlBody),
				},
				Text: &ses.Content{
					Charset: aws.String(charset),
					Data:    aws.String(rawBody),
				},
			},
			Subject: &ses.Content{
				Charset: aws.String(charset),
				Data:    aws.String(subject),
			},
		},
	}

	_, err := p.service.SendEmail(mail)
	if err != nil {
		return errors.Wrap(err, errors.ErrorUnknown, "")
	}
	return nil
}

func toAWSAddressString(p People) *string {
	addr := &mail.Address{
		Name:    p.Name,
		Address: p.Email,
	}
	return aws.String(addr.String())
}
