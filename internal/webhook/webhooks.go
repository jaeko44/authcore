package webhook

import (
	"bytes"
	"net/http"

	"authcore.io/authcore/pkg/secret"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// enumerates the events for webhooks
const (
	UpdateUserEvent = "UpdateUser"
)

// CallExternalWebhook calls an external webhook for an given event.
func CallExternalWebhook(event string, data []byte) error {
	client := &http.Client{}

	externalWebhookURL := viper.GetString("external_webhook_url")

	// No external webhook url is set. Do nothing.
	if externalWebhookURL == "" {
		return nil
	}

	externalWebhookToken := viper.Get("external_webhook_token").(secret.String).SecretString()

	req, err := http.NewRequest("POST", externalWebhookURL, bytes.NewBuffer(data))
	if err != nil {
		log.WithFields(log.Fields{
			"event": event,
			"data":  string(data),
			"err":   err,
		}).Error("cannot create new webhook request")
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Authcore-Event", event)
	if externalWebhookToken != "" {
		req.Header.Set("X-Authcore-Token", externalWebhookToken)
	}

	for trials := 1; trials <= 5; trials++ {
		_, err = client.Do(req)
		if err == nil {
			return nil
		}
	}

	log.WithFields(log.Fields{
		"event": event,
		"data":  string(data),
		"err":   err,
	}).Error("cannot call external webhook")
	return err
}
