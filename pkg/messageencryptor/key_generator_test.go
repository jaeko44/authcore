package messageencryptor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeyGenerator(t *testing.T) {
	// Mock
	keyGenerator := NewKeyGenerator([]byte("this_is_an_unbreakable_secret"))

	// Test
	key := keyGenerator.Derive("FieldEncryptor/Testing!", 32)
	assert.Len(t, key, 32)

	key2 := keyGenerator.Derive("FieldEncryptor/Testing!", 32)
	assert.Equal(t, key, key2)

	key3 := keyGenerator.Derive("FieldEncryptor/Testing?", 32)
	assert.NotEqual(t, key, key3)

	key4 := keyGenerator.Derive("FieldEncryptor/WithLongerKeyLength", 64)
	assert.Len(t, key4, 64)
}

// Ensures that the keys will not be changed when given the same secret-info pair even when
// the code is updated.
func TestKeyGeneratorCompatibility(t *testing.T) {
	// Mock
	keyGenerator := NewKeyGenerator([]byte("how_secure_could_it_be"))

	// Test
	key := keyGenerator.Derive("FieldEncryptor/ShouldBeCompetableAsWell", 32)
	expectedKey := []byte("\x63\xbf\x4f\x07\x8c\xf4\xfc\x36\x2c\x70\x13\x17\x7b\xe3\x27\xc6\x28\x25\x2d\x68\xf8\x51\xc5\xab\x59\x74\x64\x2e\xdf\x4b\x86\x42")
	assert.Equal(t, expectedKey, key)
}
