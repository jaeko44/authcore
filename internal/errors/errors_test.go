package errors

import (
	stderrors "errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
)

func TestFormat(t *testing.T) {
	err := New(ErrorNotFound, "internal error")
	assert.Equal(t, "internal error", fmt.Sprintf("%v", err))
	assert.Contains(t, fmt.Sprintf("%+v", err), "errors.TestFormat") // with stacktrace
}

func TestGRPCStatus(t *testing.T) {
	err := WithFieldViolations(ErrorNotFound, "internal error", []FieldViolation{
		FieldViolation{Field: "f", Description: "required"},
	})
	status := err.(*Error).GRPCStatus()
	assert.Equal(t, codes.NotFound, status.Code())
	assert.Equal(t, "internal error", status.Message())
}
func TestHTTPError(t *testing.T) {
	err := WithFieldViolations(ErrorNotFound, "internal error", []FieldViolation{
		FieldViolation{Field: "f", Description: "required"},
	})
	he := err.(*Error).HTTPError()
	assert.Equal(t, 404, he.Code)
	assert.Equal(t, "internal error", he.Message)
}

func TestIsKind(t *testing.T) {
	err := New(ErrorNotFound, "internal error")
	assert.True(t, IsKind(err, ErrorNotFound))

	err = stderrors.New("std error")
	assert.False(t, IsKind(err, ErrorNotFound))

	err = nil
	assert.False(t, IsKind(err, ErrorNotFound))
}

func TestAs(t *testing.T) {
	err := New(ErrorNotFound, "not found")
	err2 := &Error{}
	assert.True(t, As(err, &err2))
	assert.Equal(t, ErrorNotFound, err2.kind)
}
