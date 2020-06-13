package session

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestLoadServiceAccounts(t *testing.T) {
	viper.Set("service_accounts.testing.roles", []string{"authcore.editor"})
	viper.Set("service_accounts.testing.public_key", serviceAccountPublicKeyForTest)
	viper.Set("service_accounts.testing2.roles", []string{"authcore.admin"})
	viper.Set("service_accounts.testing2.public_key", serviceAccountPublicKeyForTest)
	serviceAccounts, err := LoadServiceAccounts()
	assert.NoError(t, err)
	assert.Len(t, serviceAccounts, 2)
	assert.Equal(t, map[string]ServiceAccount{
		"testing": ServiceAccount{
			ID:           "testing",
			PublicKeyPEM: "\n-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEHjQuqA41Mj/8B2PPb75XTeLKiacI\n0LQohjjQHORfvx3FsOWvABVP8uEZGxUWflhasFeTa/wSSp264otaxOYwFQ==\n-----END PUBLIC KEY-----\n",
			Roles:        []string{"authcore.editor"},
		},
		"testing2": ServiceAccount{
			ID:           "testing2",
			PublicKeyPEM: "\n-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEHjQuqA41Mj/8B2PPb75XTeLKiacI\n0LQohjjQHORfvx3FsOWvABVP8uEZGxUWflhasFeTa/wSSp264otaxOYwFQ==\n-----END PUBLIC KEY-----\n",
			Roles:        []string{"authcore.admin"},
		},
	}, serviceAccounts)
}
