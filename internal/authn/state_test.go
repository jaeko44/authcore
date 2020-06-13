package authn

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestValidateRedirectURI(t *testing.T) {
	viper.Set("base_url", "https://authcore.localhost/")
	viper.Set("applications.app.name", "app")
	viper.Set("applications.app.allowed_callback_urls", []string{"https://example.com"})

	assert.NoError(t, ValidateRedirectURI("app", "https://example.com/redirect"))
	assert.NoError(t, ValidateRedirectURI("app", "https://authcore.localhost/widgets/settings"))
	assert.Error(t, ValidateRedirectURI("app", "https://authcore.localhost/"))
	assert.Error(t, ValidateRedirectURI("app", "https://google.com"))
}
