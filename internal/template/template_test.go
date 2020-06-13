package template

import (
	"testing"

	"gopkg.in/go-playground/assert.v1"
)

func TestExecute(t *testing.T) {
	template := ToTemplateString("Testing {num}")
	assert.Equal(t, "Testing #1", template.Execute(map[string]string{"num": "#1"}))

	// Behavioural confirm
	template = ToTemplateString("Testing {te{test}t}")
	assert.Equal(t, "Testing {test}", template.Execute(map[string]string{"test": "s"}))

	// Undefined behavioural confirm
	template = ToTemplateString("Testing {fo{bar}}")
	assert.Equal(t, "Testing {foo}", template.Execute(map[string]string{"bar": "o", "foo": "test"}))
}
