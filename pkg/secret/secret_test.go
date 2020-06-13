package secret

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSecretString(t *testing.T) {
	secret := NewString("password")

	stringFmt := fmt.Sprintf("%v", secret)
	assert.Equal(t, "SHA256(secret):77f7205928a5c3f213d68c567049ebb7140c699287a8c4dbcf6b77934c427d5a", stringFmt)
	assert.Equal(t, "password", secret.SecretString())
	secretBytes16, err := secret.SecretBytes16()
	assert.NotNil(t, err)

	hexSecret16 := NewString("70617373776f72643132333435363738")
	stringFmt = fmt.Sprintf("%v", hexSecret16)
	secretBytes16, err = hexSecret16.SecretBytes16()
	assert.Equal(t, "SHA256(secret):391295975aebc6e68af9ad23fb61901c3525332b37411af9586cdfbbae3dca2e", stringFmt)
	assert.Nil(t, err)
	assert.Equal(t, "password12345678", string(secretBytes16[:]))

	hexSecret32 := NewString("70617373776f7264313233343536373870617373776f72643132333435363738")
	secretBytes16, err = hexSecret32.SecretBytes16()
	assert.NotNil(t, err)
	secretBytes32, err := hexSecret32.SecretBytes32()
	assert.Nil(t, err)
	assert.Equal(t, "password12345678password12345678", string(secretBytes32[:]))
}
