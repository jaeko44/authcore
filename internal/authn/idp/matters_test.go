package idp

import (
	"testing"

	"authcore.io/authcore/internal/config"
	"authcore.io/authcore/pkg/secret"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestMattersIDP(t *testing.T) {
	config.InitDefaults()
	viper.Set("matters_app_id", "testing")
	viper.Set("matters_app_secret", secret.NewString("testing"))
	viper.Set("matters_url", "https://server-stage.matters.news")
	viper.Set("matters_id_token_certificate", `
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwAGlzdzVZeaph74lpMz4
kd+cKhQSCXm7YBzQ2ZwJG38rKb1xd0nGYERULT7g75RPcsBJI2JmcpTty45KMDXn
qGzxZcvdrut7hqe2bjR5rSBGM/YuEx3fkSUTT+fQsodFfbSscah876JfndppVHog
C5zQ011k04wkua8OXzdMb0NwzOndUVJp16u9g1NGsUVxayvi1J+3SvUE4jtkZJtF
bH6eZOTGMwbqakLebDDuydYfe1SPBOYLvy++jvVwFxThv6tOAL2iJ0S5RXOaTFfe
e9gBNPFwHWh4HMKCMggbCooLirxswMuXI5tcNcwr8tJdSSK5nwxyOAub0q+OAmhu
3wIDAQAB
-----END PUBLIC KEY-----
	`)

	defer viper.Reset()
	provider := NewMattersIDP()
	url, _, err := provider.AuthorizationURL("testing")
	assert.Equal(t, "matters", provider.ID())
	assert.NoError(t, err)
	assert.Equal(t, "https://server-stage.matters.news/oauth/authorize?client_id=testing&prompt=select_account&redirect_uri=https%3A%2F%2Fauthcore.localhost%2Foauth%2Fredirect&response_type=code&scope=query%3Aviewer%3Ainfo%3Aemail&state=testing", url)
}
