package idp

import (
	"testing"

	"authcore.io/authcore/internal/config"
	"authcore.io/authcore/pkg/secret"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestAppleIDP(t *testing.T) {
	config.InitDefaults()
	viper.Set("app_client_id", "testing")
	viper.Set("apple_app_key_id", "testing")
	viper.Set("apple_app_key_issuer", "testing")
	viper.Set("apple_app_private_key", secret.NewString(`
-----BEGIN PRIVATE KEY-----
MIGTAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBHkwdwIBAQQgcBw7MsBZhJGWNrnF
zFxMbRcTgOmcV1PjZ0b+wbj8suegCgYIKoZIzj0DAQehRANCAAQgHCdhuoQHss3S
QWu9cKRyBh+bp7aTIY/yaWS/Df44n3OKjp8g0yto24IuSSREc24bFLBwCApUIAog
HhYGjVDE
-----END PRIVATE KEY-----
	`))
	defer viper.Reset()
	provider := NewAppleIDP()
	url, _, err := provider.AuthorizationURL("testing")
	assert.Equal(t, "apple", provider.ID())
	assert.NoError(t, err)
	assert.Equal(t, "https://appleid.apple.com/auth/authorize?client_id=&redirect_uri=https%3A%2F%2Fauthcore.localhost%2Foauth%2Fredirect&response_mode=form_post&response_type=code&scope=email&state=testing", url)
}
