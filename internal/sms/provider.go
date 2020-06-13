package sms

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/internal/template"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/text/language"
)

// sendSMS sends a SMS
func sendSMS(contentTemplate template.String, contentMap map[string]string, phone string) error {
	contentMap["application_name"] = viper.GetString("application_name")
	message := contentTemplate.Execute(contentMap)

	if strings.HasSuffix(os.Args[0], ".test") {
		log.WithFields(log.Fields{
			"to":      phone,
			"message": message,
		}).Infof("skip sending SMS in unit tests")
		return nil
	}

	accountSid := viper.GetString("twilio_account_sid")
	authToken := viper.GetString("twilio_auth_token")

	urlStr := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", accountSid)

	msgData := url.Values{}
	msgData.Set("MessagingServiceSid", viper.GetString("twilio_service_sid"))
	msgData.Set("To", phone)
	msgData.Set("Body", message)
	msgDataReader := *strings.NewReader(msgData.Encode())

	client := &http.Client{}
	req, _ := http.NewRequest("POST", urlStr, &msgDataReader)
	req.SetBasicAuth(accountSid, authToken)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	_, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, errors.ErrorUnknown, "")
	}
	return nil
}

func defaultLanguage() (language.Tag, error) {
	defaultLanguage := viper.GetString("default_language")
	tag, err := language.Parse(defaultLanguage)
	if err != nil {
		return language.Tag{}, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	return tag, nil
}
