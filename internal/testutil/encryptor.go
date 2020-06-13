package testutil

import (
	"authcore.io/authcore/pkg/messageencryptor"
	"authcore.io/authcore/pkg/secret"

	"github.com/spf13/viper"
)

// EncryptorForTest creates a *MessageEncryptor for tests.
func EncryptorForTest() *messageencryptor.MessageEncryptor {
	secret, _ := viper.Get("secret_key_base").(secret.String).SecretBytes()
	keyGenerator := messageencryptor.NewKeyGenerator(secret)
	encryptor, _ := messageencryptor.NewMessageEncryptor(
		keyGenerator.Derive(
			"FieldEncryptor/Xsalsa20Poly1305",
			messageencryptor.CipherXsalsa20Poly1305.KeyLength(),
		),
		messageencryptor.CipherXsalsa20Poly1305,
	)
	return encryptor
}
