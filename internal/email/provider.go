package email

import (
	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/pkg/secret"

	"github.com/spf13/viper"
)

// Provider is an interface implementing Send.
type Provider interface {
	Send(from, to People, subject, rawBody, htmlBody string) error
}

// People is a struct that defines an user: a sender or a receipient.
type People struct {
	Name  string
	Email string
}

func getProvider() (Provider, error) {
	emailProvider := viper.GetString("email_delegate")
	if emailProvider == "sendgrid" {
		sendgridAPIKey := viper.Get("sendgrid_api_key").(secret.String).SecretString()
		if sendgridAPIKey == "" {
			return nil, errors.New(errors.ErrorUnknown, "cannot get a set of valid confidential for sendgrid email delegate")
		}
		return NewSendgridEmailProvider(sendgridAPIKey)
	} else if emailProvider == "ses" {
		awsRegion := viper.GetString("aws_ses_region")
		awsAccessKeyID := viper.GetString("aws_ses_access_key_id")
		awsSecretAccessKey := viper.Get("aws_ses_secret_access_key").(secret.String).SecretString()
		if awsRegion == "" || awsAccessKeyID == "" || awsSecretAccessKey == "" {
			return nil, errors.New(errors.ErrorUnknown, "cannot get a set of valid confidential for ses email delegate")
		}
		return NewSESEmailProvider(awsRegion, awsAccessKeyID, awsSecretAccessKey)
	}
	return nil, errors.New(errors.ErrorUnknown, "an unknown email delegate is defined")
}
