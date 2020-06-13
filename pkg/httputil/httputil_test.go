package httputil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetIPAddressFromXFF(t *testing.T) {
	assert.Equal(t, GetIPAddrFromXFF("127.0.0.1"), "127.0.0.1")
	assert.Equal(t, GetIPAddrFromXFF("127.0.0.1,127.0.0.2"), "127.0.0.2")
	assert.Equal(t, GetIPAddrFromXFF("127.0.0.1, 127.0.0.2"), "127.0.0.2")
}

func TestNormalizeURI(t *testing.T) {
	uri, err := NormalizeURI("https://google.com/")
	assert.NoError(t, err)
	assert.Equal(t, "https://google.com/", uri)

	uri, err = NormalizeURI("https://people.blocksq.com/~dummy/../~alice")
	assert.NoError(t, err)
	assert.Equal(t, "https://people.blocksq.com/~alice", uri)

	uri, err = NormalizeURI("https://people.blocksq.com/~alice/./secret")
	assert.NoError(t, err)
	assert.Equal(t, "https://people.blocksq.com/~alice/secret", uri)
}
