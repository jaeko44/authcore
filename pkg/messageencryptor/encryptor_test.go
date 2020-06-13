package messageencryptor

import (
	"bytes"
	"testing"

	"authcore.io/authcore/pkg/nulls"

	"github.com/stretchr/testify/assert"
)

func TestMessageEncryptor(t *testing.T) {
	// Mock
	key := []byte("\x63\xbf\x4f\x07\x8c\xf4\xfc\x36\x2c\x70\x13\x17\x7b\xe3\x27\xc6\x28\x25\x2d\x68\xf8\x51\xc5\xab\x59\x74\x64\x2e\xdf\x4b\x86\x42")
	messageEncryptor, err := NewMessageEncryptor(key, CipherXsalsa20Poly1305)
	assert.NoError(t, err)

	// Test
	ciphertext, err := messageEncryptor.Encrypt([]byte("hello world"), []byte("additional_data"))
	assert.NoError(t, err)

	plaintext, err := messageEncryptor.Decrypt(ciphertext, []byte("additional_data"))
	assert.NoError(t, err)
	assert.Equal(t, []byte("hello world"), plaintext)

	_, err = messageEncryptor.Decrypt(ciphertext, []byte("incorrect_additional_data"))
	assert.Error(t, err)

	type StructTest struct {
		StringField                 string `encrypt:"" encryptPurpose:"Test"`
		EncryptedStringField        nulls.String
		NullStringField             nulls.String `encrypt:"OtherEncryptedField"`
		OtherEncryptedField         string
		NullByteSliceField          nulls.ByteSlice `encrypt:""`
		EncryptedNullByteSliceField string
	}

	type NestedStructTest struct {
		Nested StructTest `encrypt:"-"`
	}

	s := StructTest{
		StringField:        "test1",
		NullStringField:    nulls.String{String: "test2", Valid: true},
		NullByteSliceField: nulls.ByteSlice{ByteSlice: []byte{0x01, 0x02}, Valid: true},
	}

	err = messageEncryptor.EncryptStruct(&s)
	if assert.NoError(t, err) {
		assert.True(t, s.EncryptedStringField.Valid)
		assert.NotEmpty(t, s.OtherEncryptedField)
		assert.NotEmpty(t, s.EncryptedNullByteSliceField)
	}

	s.StringField = ""
	s.NullStringField = nulls.String{}
	s.NullByteSliceField = nulls.ByteSlice{}

	err = messageEncryptor.DecryptStruct(&s)
	if assert.NoError(t, err) {
		assert.Equal(t, "test1", s.StringField)
		assert.True(t, s.NullStringField.Valid)
		assert.Equal(t, "test2", s.NullStringField.String)
		assert.True(t, s.NullByteSliceField.Valid)
		assert.True(t, bytes.Equal(s.NullByteSliceField.ByteSlice, []byte{0x01, 0x02}))
	}

	oldValue1 := s.EncryptedStringField
	oldValue2 := s.OtherEncryptedField
	s.NullStringField.Valid = true
	s.NullStringField.String = "test3"

	// Unchanged field should not change ciphertext field
	err = messageEncryptor.EncryptStruct(&s)
	if assert.NoError(t, err) {
		assert.Equal(t, oldValue1, s.EncryptedStringField)
		assert.NotEqual(t, oldValue2, s.OtherEncryptedField)
	}

	// Test nested struct

	ns := NestedStructTest{
		Nested: StructTest{
			StringField:     "test3",
			NullStringField: nulls.String{String: "test4", Valid: true},
		},
	}
	err = messageEncryptor.EncryptStruct(&ns)
	if assert.NoError(t, err) {
		assert.True(t, ns.Nested.EncryptedStringField.Valid)
		assert.NotEmpty(t, ns.Nested.EncryptedStringField.String)
		assert.NotEmpty(t, ns.Nested.OtherEncryptedField)
	}
}

// Ensures that the encryption algorithm could decrypt previously-encrypted messages even if
// there are code updates.
func TestMessageEncryptorCompatibility(t *testing.T) {
	// 1. xsalsa20-poly1305
	key := []byte("\x63\xbf\x4f\x07\x8c\xf4\xfc\x36\x2c\x70\x13\x17\x7b\xe3\x27\xc6\x28\x25\x2d\x68\xf8\x51\xc5\xab\x59\x74\x64\x2e\xdf\x4b\x86\x42")
	messageEncryptor, err := NewMessageEncryptor(key, CipherXsalsa20Poly1305)
	assert.NoError(t, err)

	plaintext, err := messageEncryptor.Decrypt(
		"DoT8eQx2op7fzgCpM7XrMik58mTos1urkIGqC_NGLoKIBwtRh-G_gplL7CcZ3c080L8L8_cEh88J7VUTHqz33WH9JjmRFFc",
		[]byte("additional_data"),
	)
	assert.NoError(t, err)
	assert.Equal(t, []byte("hello world"), plaintext)
}

func TestMessageEncryptorWithOutdatedKeys(t *testing.T) {
	// Mock
	oldKeyGenerator := NewKeyGenerator([]byte("this_key_should_be_very_secure_or_is_it"))
	oldKey := oldKeyGenerator.Derive("this_is_an_outdated_key", CipherXsalsa20Poly1305.KeyLength())

	newKeyGenerator := NewKeyGenerator([]byte("oops_the_previous_key_is_leaked_so_we_updated_that"))
	newKey := newKeyGenerator.Derive("this_is_the_latest_key", CipherXsalsa20Poly1305.KeyLength())

	oldMessageEncryptor, err := NewMessageEncryptor(oldKey, CipherXsalsa20Poly1305)
	assert.NoError(t, err)
	newMessageEncryptor, err := NewMessageEncryptor(newKey, CipherXsalsa20Poly1305)
	assert.NoError(t, err)
	newMessageEncryptor.AddOldKey(oldKey, CipherXsalsa20Poly1305)

	// Test
	// 1. Encrypt with old message encryptor
	ciphertext, err := oldMessageEncryptor.Encrypt([]byte("hello world"), []byte("additional_data"))
	assert.NoError(t, err)

	// 2. Decrypt with new message encryptor
	plaintext, err := newMessageEncryptor.Decrypt(ciphertext, []byte("additional_data"))
	assert.NoError(t, err)
	assert.Equal(t, []byte("hello world"), plaintext)
}
