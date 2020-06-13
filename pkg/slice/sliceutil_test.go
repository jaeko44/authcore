package slice

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIndex(t *testing.T) {
	a := []int{1243, 12, 51, 241, 789, 383, 92498}
	assert.Equal(t, 2, Index(a, 51))
	assert.Equal(t, 0, Index(a, 1243))
	assert.Equal(t, -1, Index(a, 63))
}

func TestContains(t *testing.T) {
	a := []int{1243, 12, 51, 241, 789, 383, 92498}
	assert.True(t, Contains(a, 51))
	assert.True(t, Contains(a, 1243))
	assert.False(t, Contains(a, 63))
}
