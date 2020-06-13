package session

import (
	"crypto/ecdsa"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/spf13/viper"
)

const (
	// ServiceAccountPrefix is a prefix for the issuer field in service account token.
	ServiceAccountPrefix string = "serviceaccount:"
)

// ServiceAccount represents a service account.
type ServiceAccount struct {
	ID           string
	PublicKeyPEM string `mapstructure:"public_key"`
	Roles        []string
}

// KeyID returns the JWT key ID.
func (a *ServiceAccount) KeyID() (string, error) {
	k, err := a.PublicKey()
	if err != nil {
		return "", err
	}
	return kidFromECPublicKey(k)
}

// PublicKey returns the JWT public key of the service account.
func (a *ServiceAccount) PublicKey() (*ecdsa.PublicKey, error) {
	return jwt.ParseECPublicKeyFromPEM([]byte(a.PublicKeyPEM))
}

// SubjectString is a string that represents the account.
func (a *ServiceAccount) SubjectString() string {
	return ServiceAccountPrefix + a.ID
}

// HasRole returns whether the serviec account has the given role.
func (a *ServiceAccount) HasRole(role string) bool {
	for _, r := range a.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// LoadServiceAccounts loads service accounts from config files.
func LoadServiceAccounts() (accounts map[string]ServiceAccount, err error) {
	rawMap := make(map[string]ServiceAccount, 0)
	err = viper.UnmarshalKey("service_accounts", &rawMap)
	if err != nil {
		return nil, err
	}

	accounts = make(map[string]ServiceAccount)
	for k, v := range rawMap {
		v.ID = strings.ToLower(k)
		accounts[v.ID] = v
	}
	return
}
