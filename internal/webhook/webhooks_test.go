package webhook

import (
	"io/ioutil"
	"net/http"
	"testing"

	"authcore.io/authcore/pkg/secret"

	"github.com/jarcoal/httpmock"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestWebhook(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// Mocking
	httpmock.RegisterResponder("POST", "https://authcore.dev/callback/authcore",
		func(req *http.Request) (*http.Response, error) {
			resp, err := httpmock.NewJsonResponse(200, map[string]interface{}{})
			if err != nil {
				return httpmock.NewStringResponse(500, ""), nil
			}
			defer resp.Body.Close()

			assert.Equal(t, "", req.Header.Get("X-Authcore-Token"))
			assert.Equal(t, "Test", req.Header.Get("X-Authcore-Event"))

			body, err := ioutil.ReadAll(req.Body)
			assert.NoError(t, err)
			assert.Equal(t, []byte("{\"foo\":\"bar\"}"), body)

			return resp, nil
		},
	)

	viper.Set("external_webhook_url", "https://authcore.dev/callback/authcore")
	viper.Set("external_webhook_token", secret.NewString(""))

	err := CallExternalWebhook("Test", []byte("{\"foo\":\"bar\"}"))
	assert.NoError(t, err)
}

func TestWebhookWithToken(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// Mocking
	httpmock.RegisterResponder("POST", "https://authcore.dev/callback/authcore",
		func(req *http.Request) (*http.Response, error) {
			resp, err := httpmock.NewJsonResponse(200, map[string]interface{}{})
			if err != nil {
				return httpmock.NewStringResponse(500, ""), nil
			}
			defer resp.Body.Close()

			assert.Equal(t, "LVRlc3RIZWFkZXIt", req.Header.Get("X-Authcore-Token"))
			assert.Equal(t, "Test", req.Header.Get("X-Authcore-Event"))

			body, err := ioutil.ReadAll(req.Body)
			assert.NoError(t, err)
			assert.Equal(t, []byte("{\"foo\":\"bar\"}"), body)

			return resp, nil
		},
	)

	viper.Set("external_webhook_url", "https://authcore.dev/callback/authcore")
	viper.Set("external_webhook_token", secret.NewString("LVRlc3RIZWFkZXIt"))

	err := CallExternalWebhook("Test", []byte("{\"foo\":\"bar\"}"))
	assert.NoError(t, err)
}
