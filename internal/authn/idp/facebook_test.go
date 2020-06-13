package idp

import (
	"testing"

	"authcore.io/authcore/internal/config"
	"authcore.io/authcore/pkg/secret"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestFacebookIDP(t *testing.T) {
	config.InitDefaults()
	viper.Set("facebook_app_id", "testing")
	viper.Set("facebook_app_secret", secret.NewString("testing"))
	defer viper.Reset()
	provider := NewFacebookIDP()
	url, _, err := provider.AuthorizationURL("testing")
	assert.Equal(t, "facebook", provider.ID())
	assert.NoError(t, err)
	assert.Equal(t, "https://www.facebook.com/v3.2/dialog/oauth?client_id=testing&display=popup&redirect_uri=https%3A%2F%2Fauthcore.localhost%2Foauth%2Fredirect&response_type=code&scope=email&state=testing", url)
}
