package messageencryptor

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"reflect"

	"authcore.io/authcore/pkg/nulls"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/vmihailenco/msgpack"
	"golang.org/x/crypto/nacl/secretbox"
)

// Cipher is a type enumerating the algorithms for the encryptor.
type Cipher int32

// Enunmerates the encryptor algorithms.
const (
	CipherXsalsa20Poly1305 Cipher = 0
)

var nullStringType reflect.Type
var nullByteSliceType reflect.Type

func init() {
	nullStringType = reflect.TypeOf(nulls.String{})
	nullByteSliceType = reflect.TypeOf(nulls.ByteSlice{})
}

// Key represents a key - it involves the secret key and the cipher used for encryption.
type Key struct {
	key    []byte
	cipher Cipher
}

// MessageEncryptor represents a message encryptor.
type MessageEncryptor struct {
	keys []Key
}

// NewMessageEncryptor returns a new message encryptor.
func NewMessageEncryptor(key []byte, cipher Cipher) (*MessageEncryptor, error) {
	if len(key) != cipher.KeyLength() {
		return nil, errors.Errorf("key length for message encryptor mismatch: expected %d, given %d", cipher.KeyLength(), len(key))
	}
	// generate key from keyGenerator. specify length from cipher.
	return &MessageEncryptor{
		keys: []Key{
			Key{
				key:    key,
				cipher: cipher,
			},
		},
	}, nil
}

// AddOldKey adds an old key to the current encryptor.
func (messageEncryptor *MessageEncryptor) AddOldKey(key []byte, cipher Cipher) {
	messageEncryptor.keys = append(
		messageEncryptor.keys,
		Key{
			key:    key,
			cipher: cipher,
		},
	)
}

// Encrypt encrypts a message with the current key.  Outputs in base64.
func (messageEncryptor *MessageEncryptor) Encrypt(plaintext, purpose []byte) (string, error) {
	key := messageEncryptor.keys[0]
	return encrypt(plaintext, purpose, key)
}

// Decrypt decrypts a ciphertext with the latest key to the older keys until it is correctly decrypted.
func (messageEncryptor *MessageEncryptor) Decrypt(ciphertext string, purpose []byte) ([]byte, error) {
	var plaintext []byte
	var err error
	for _, key := range messageEncryptor.keys {
		plaintext, err = decrypt(ciphertext, purpose, key)
		if err == nil {
			break
		}
	}
	if err != nil {
		return []byte{}, errors.New("decrypt failed")
	}
	return plaintext, err
}

// EncryptStruct encrypts fields in a struct as specified by struct tags. Two struct tags are defined:
//
// `encrypt:"FieldName"` encrypts the field and saved the encrypted text to `FieldName`. If
// `FieldName` is empty, "Encrypted"+FieldName will be used.
//
// `encrypt:"-"` encrypts the nested struct fields.
func (messageEncryptor *MessageEncryptor) EncryptStruct(s interface{}) error {
	return forEachTaggedFields(s, func(src, tgt reflect.Value, purpose []byte) error {
		pt, srcValid := getByteSliceValue(src)
		if srcValid {
			oldCt, tgtValid := getStringValue(tgt)
			if tgtValid {
				oldPt, err := messageEncryptor.Decrypt(oldCt, purpose)
				if err == nil && bytes.Equal(oldPt, pt) {
					return nil // skip unchanged values
				}
			}

			ct, err := messageEncryptor.Encrypt([]byte(pt), purpose)
			if err != nil {
				return err
			}
			setStringValue(tgt, ct, true)
		} else {
			setStringValue(tgt, "", false)
		}
		return nil
	})
}

// DecryptStruct decrypts fields in a struct as specified by the `encrypt` struct tags
func (messageEncryptor *MessageEncryptor) DecryptStruct(s interface{}) error {
	return forEachTaggedFields(s, func(src, tgt reflect.Value, purpose []byte) error {
		ct, valid := getStringValue(tgt)
		if valid {
			pt, err := messageEncryptor.Decrypt(ct, purpose)
			if err != nil {
				return err
			}
			setByteSliceValue(src, pt, true)
		} else {
			setByteSliceValue(src, []byte{}, false)
		}
		return nil
	})
}

// KeyLength returns the key length of a given cipher.
func (cipher Cipher) KeyLength() int {
	switch cipher {
	case CipherXsalsa20Poly1305:
		return 32
	}
	log.WithFields(log.Fields{
		"cipher": cipher,
	}).Fatal("key length for cipher is not defined")
	return 0
}

// encryption and decryption bundling

func encrypt(plaintext, purpose []byte, key Key) (string, error) {
	var bCiphertext []byte
	var err error
	switch key.cipher {
	case CipherXsalsa20Poly1305:
		bCiphertext, err = encryptXsalsa20Poly1305(plaintext, purpose, key.key)
	}

	if err != nil {
		return "", err
	}
	ciphertext := base64.RawURLEncoding.EncodeToString(bCiphertext)
	return ciphertext, nil
}

func decrypt(ciphertext string, purpose []byte, key Key) ([]byte, error) {
	bCiphertext, err := base64.RawURLEncoding.DecodeString(ciphertext)
	if err != nil {
		return []byte{}, err
	}
	var bPlaintext []byte
	switch key.cipher {
	case CipherXsalsa20Poly1305:
		bPlaintext, err = decryptXsalsa20Poly1305(bCiphertext, purpose, key.key)
	}
	if err != nil {
		return []byte{}, err
	}
	return bPlaintext, nil
}

// encryption and decryption for xsalsa20-poly1305

func encryptXsalsa20Poly1305(plaintext, purpose, vKey []byte) ([]byte, error) {
	var nonce [24]byte
	var key [32]byte
	_, err := rand.Read(nonce[:])
	if err != nil {
		return []byte{}, err
	}
	copy(key[:], vKey)

	packedPlaintext, err := msgpack.Marshal([][]byte{plaintext, purpose})
	if err != nil {
		return []byte{}, err
	}

	ciphertext := secretbox.Seal(nonce[:], packedPlaintext, &nonce, &key)
	return ciphertext, nil
}

func decryptXsalsa20Poly1305(ciphertext, purpose, vKey []byte) ([]byte, error) {
	var nonce [24]byte
	var key [32]byte
	copy(nonce[:], ciphertext[:24])
	copy(key[:], vKey)

	packedPlaintext, ok := secretbox.Open(nil, ciphertext[24:], &nonce, &key)
	if !ok {
		return []byte{}, errors.New("decrypt failed")
	}

	var unpackedPlaintext [][]byte
	err := msgpack.Unmarshal(packedPlaintext, &unpackedPlaintext)
	if err != nil {
		return []byte{}, err
	}
	if len(unpackedPlaintext) < 2 {
		return []byte{}, errors.New("insufficient entries to unpack")
	}
	plaintext := unpackedPlaintext[0]
	decryptedPurpose := unpackedPlaintext[1]

	if !bytes.Equal(purpose, decryptedPurpose) {
		return []byte{}, errors.New("unmatched additional data")
	}

	return plaintext, nil
}

func forEachTaggedFields(s interface{}, op func(src, target reflect.Value, purpose []byte) error) error {
	ptr := reflect.ValueOf(s)
	if ptr.Kind() != reflect.Ptr {
		return errors.Errorf("%v should be a pointer of struct. Got %v", s, ptr.Kind())
	}
	v := ptr.Elem()
	t := v.Type()
	if t.Kind() != reflect.Struct {
		return errors.Errorf("%v should be a struct. Got %v", s, t.Kind())
	}

	for i := 0; i < t.NumField(); i++ {
		srcField := t.Field(i)
		tgtName, ok := srcField.Tag.Lookup("encrypt")
		if ok {
			// Inner struct
			if tgtName == "-" && srcField.Type.Kind() == reflect.Struct {
				vv := v.FieldByName(srcField.Name)
				s = vv.Addr().Interface()
				err := forEachTaggedFields(s, op)
				if err != nil {
					return err
				}
				continue
			}

			if tgtName == "" {
				tgtName = "Encrypted" + srcField.Name
			}
			tgtField, ok := t.FieldByName(tgtName)
			if !ok {
				return errors.Errorf("cannot encrypt struct: missing field %v", tgtName)
			}
			if srcField.Type.Kind() != reflect.String && !srcField.Type.ConvertibleTo(nullStringType) &&
				!srcField.Type.ConvertibleTo(nullByteSliceType) {
				return errors.Errorf("cannot encrypt struct: %v is not string or []byte type", srcField.Name)
			}
			if tgtField.Type.Kind() != reflect.String && !tgtField.Type.ConvertibleTo(nullStringType) {
				return errors.Errorf("cannot encrypt struct: %v is not string type", tgtField.Name)
			}
			srcValue := v.FieldByName(srcField.Name)
			tgtValue := v.FieldByName(tgtField.Name)
			purpose, ok := srcField.Tag.Lookup("encryptPurpose")
			if !ok {
				purpose = fmt.Sprintf("struct:%v.%v", t.String(), tgtField.Name)
			}
			err := op(srcValue, tgtValue, []byte(purpose))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func getStringValue(v reflect.Value) (value string, valid bool) {
	valid = false
	value = ""
	if v.Type().ConvertibleTo(nullStringType) {
		valid = v.FieldByName("Valid").Bool()
		value = v.FieldByName("String").String()
	} else if v.Kind() == reflect.String {
		value = v.String()
		valid = len(value) > 0
	} else {
		log.Fatalf("MessageEncryptor does not support reading %v type", v.Kind())
	}
	return
}

func setStringValue(v reflect.Value, s string, valid bool) {
	if v.Type().ConvertibleTo(nullStringType) {
		v.FieldByName("Valid").SetBool(valid)
		v.FieldByName("String").SetString(s)
	} else if v.Kind() == reflect.String {
		if valid {
			v.SetString(s)
		} else {
			v.SetString("")
		}
	} else {
		log.Fatalf("MessageEncryptor does not support writing %v type", v.Kind())
	}
}

func getByteSliceValue(v reflect.Value) (value []byte, valid bool) {
	valid = false
	value = []byte{}
	if v.Type().ConvertibleTo(nullByteSliceType) {
		valid = v.FieldByName("Valid").Bool()
		value = []byte(v.FieldByName("ByteSlice").Bytes())
	} else if v.Type().ConvertibleTo(nullStringType) {
		valid = v.FieldByName("Valid").Bool()
		value = []byte(v.FieldByName("String").String())
	} else if v.Kind() == reflect.String {
		value = []byte(v.String())
		valid = len(value) > 0
	}
	return
}

func setByteSliceValue(v reflect.Value, s []byte, valid bool) {
	if v.Type().ConvertibleTo(nullByteSliceType) {
		v.FieldByName("Valid").SetBool(valid)
		v.FieldByName("ByteSlice").SetBytes(s)
	} else if v.Type().ConvertibleTo(nullStringType) {
		v.FieldByName("Valid").SetBool(valid)
		v.FieldByName("String").SetString(string(s))
	} else if v.Kind() == reflect.String {
		if valid {
			v.SetString(string(s))
		} else {
			v.SetString("")
		}
	}
}
