package clientapp

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestGetByClientID(t *testing.T) {
	viper.Set("applications.testing.name", "Testing")
	viper.Set("applications.testing.app_domains", []string{"authcore.testing"})
	viper.Set("applications.testing.allowed_callback_urls", []string{"https://authcore.testing"})
	viper.Set("applications.testing.idp_list", []string{
		"google",
		"facebook",
	})
	clientApp, err := GetByClientID("testing")
	if assert.NoError(t, err) {
		assert.Equal(t, "testing", clientApp.ID)
		assert.Equal(t, "Testing", clientApp.Name)
		assert.Equal(t, []string{"authcore.testing"}, clientApp.AppDomains)
		assert.Equal(t, []string{"https://authcore.testing"}, clientApp.AllowedCallbackURLs)
		assert.Equal(t, []string{
			"google",
			"facebook",
		}, clientApp.IDPList)
	}

	clientApp, err = GetByClientID("nonexist")
	if assert.Error(t, err) {
		assert.Nil(t, clientApp)
	}

	viper.Set("default_client_id", "testing")
	clientApp, err = GetByClientID("")
	if assert.NoError(t, err) {
		assert.Equal(t, "testing", clientApp.ID)
		assert.Equal(t, "Testing", clientApp.Name)
		assert.Equal(t, []string{"authcore.testing"}, clientApp.AppDomains)
		assert.Equal(t, []string{"https://authcore.testing"}, clientApp.AllowedCallbackURLs)
		assert.Equal(t, []string{
			"google",
			"facebook",
		}, clientApp.IDPList)
	}

	// test for "authcore.io" client id
	clientApp, err = GetByClientID("authcore.io")
	if assert.NoError(t, err) {
		assert.Equal(t, "testing", clientApp.ID)
		assert.Equal(t, "Testing", clientApp.Name)
		assert.Equal(t, []string{"authcore.testing"}, clientApp.AppDomains)
		assert.Equal(t, []string{"https://authcore.testing"}, clientApp.AllowedCallbackURLs)
		assert.Equal(t, []string{
			"google",
			"facebook",
		}, clientApp.IDPList)
	}
}
