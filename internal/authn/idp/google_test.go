package idp

import (
	"testing"

	"authcore.io/authcore/internal/config"
	"authcore.io/authcore/pkg/secret"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestGoogleIDP(t *testing.T) {
	config.InitDefaults()
	viper.Set("google_app_id", "testing")
	viper.Set("google_app_secret", secret.NewString("testing"))
	defer viper.Reset()
	provider := NewGoogleIDP()
	url, _, err := provider.AuthorizationURL("testing")
	assert.Equal(t, "google", provider.ID())
	assert.NoError(t, err)
	assert.Equal(t, "https://accounts.google.com/o/oauth2/auth?client_id=testing&prompt=select_account&redirect_uri=https%3A%2F%2Fauthcore.localhost%2Foauth%2Fredirect&response_type=code&scope=email&state=testing", url)
}
