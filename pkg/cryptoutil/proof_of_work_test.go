package cryptoutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVerifyProofOfWork(t *testing.T) {
	// Valid PoW
	{
		ValidPow, err := VerifyProofOfWork("nOqvUGaUD5BtG2B74YI4YQ", "L1rPhkbhSx9wYiT9r6emvw", 65536)
		assert.True(t, ValidPow)
		assert.Nil(t, err)
	}
	{
		ValidPow, err := VerifyProofOfWork("pQ3AVAEZTxu3AF4ZJEtHpg", "v2lB6snRkPDSLMZq3H_j9A", 385644155)
		assert.True(t, ValidPow)
		assert.Nil(t, err)
	}
	// Invalid PoW
	{
		ValidPow, err := VerifyProofOfWork("pQ3AVAEZTxu3AF4ZJEtHpg", "v2lB6snRkPDSLMZq3H_j9A", 385644156)
		assert.False(t, ValidPow)
		assert.NotNil(t, err)
	}
}

func TestSolveProofOfWork(t *testing.T) {
	proof, err := SolveProofOfWork("nOqvUGaUD5BtG2B74YI4YQ", 1024)
	t.Log(proof)
	assert.Nil(t, err)
	validPow, err := VerifyProofOfWork("nOqvUGaUD5BtG2B74YI4YQ", proof, 1024)
	assert.Nil(t, err)
	assert.True(t, validPow)
}
