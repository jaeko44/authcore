package email

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToAWSAddressString(t *testing.T) {
	p := People{
		Name:  "Alice",
		Email: "alice@example.com",
	}
	assert.Equal(t, "\"Alice\" <alice@example.com>", *toAWSAddressString(p))

	p = People{
		Name:  "Mallory <mallory@example.com>",
		Email: "alice@example.com",
	}
	assert.Equal(t, "\"Mallory <mallory@example.com>\" <alice@example.com>", *toAWSAddressString(p))

	p = People{
		Name:  "Mallory <mallory@example.com>\"",
		Email: "alice@example.com",
	}
	assert.Equal(t, "\"Mallory <mallory@example.com>\\\"\" <alice@example.com>", *toAWSAddressString(p))

	p = People{
		Name:  "埃德溫",
		Email: "edwin@example.com",
	}
	assert.Equal(t, "=?utf-8?q?=E5=9F=83=E5=BE=B7=E6=BA=AB?= <edwin@example.com>", *toAWSAddressString(p))

	p = People{
		Name:  "New\nLine",
		Email: "newline@example.com",
	}
	assert.Equal(t, "=?utf-8?q?New=0ALine?= <newline@example.com>", *toAWSAddressString(p))
}
