package errors

import (
	"database/sql"
	stderrors "errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	stringsUtil "authcore.io/authcore/pkg/strings"

	"github.com/go-sql-driver/mysql"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"github.com/go-playground/validator/v10"
)

// Kind is the kind of error.
type Kind string

// Error kinds
const (
	ErrorCanceled               Kind = "canceled"
	ErrorUnknown                Kind = "unknown error"
	ErrorInvalidArgument        Kind = "invalid argument"
	ErrorDeadlineExceeded       Kind = "deadline exceeded"
	ErrorNotFound               Kind = "entity not found"
	ErrorAlreadyExists          Kind = "already exists"
	ErrorPermissionDenied       Kind = "permission denied"
	ErrorUnauthenticated        Kind = "unauthenticated"
	ErrorResourceExhausted      Kind = "too many requests"
	ErrorFailedPrecondition     Kind = "failed precondition"
	ErrorAborted                Kind = "operation aborted"
	ErrorUnavailable            Kind = "service unavailable"
	ErrorUserTemporarilyBlocked Kind = "user has been temporarily blocked"
)

// FieldViolation is a struct for providing field error details in HTTP error. It matches the same struct in errdetails package
type FieldViolation struct {
	Field       string
	Description string
}

// Error is an internal errors with stacktrace. It can be converted to a HTTP response or a GRPC
// status.
type Error struct {
	error
	kind            Kind
	fieldViolations []FieldViolation
}

// Format formats the error.
func (e *Error) Format(s fmt.State, verb rune) {
	if formatter, ok := e.error.(fmt.Formatter); ok {
		formatter.Format(s, verb)
		return
	}
	io.WriteString(s, e.Error())
}

// Kind returns the error kind.
func (e *Error) Kind() Kind {
	return e.kind
}

// FieldViolations returns a structure that represents field validation errors.
func (e *Error) FieldViolations() []FieldViolation {
	return e.fieldViolations
}

// GRPCStatus implements the GRPCStatus from internal error, it builds field violations detail if provided.
func (e *Error) GRPCStatus() *status.Status {
	code := GRPCStatusCodeFromKind(e.kind)
	status := status.New(code, e.Error())
	// Build field violations if provided
	if len(e.fieldViolations) > 0 {
		br := &errdetails.BadRequest{}
		for _, fieldViolation := range e.fieldViolations {
			br.FieldViolations = append(br.FieldViolations, &errdetails.BadRequest_FieldViolation{
				Field:       fieldViolation.Field,
				Description: fieldViolation.Description,
			})
		}
		status, _ = status.WithDetails(br)
	}
	return status
}

// HTTPError converts an Error into HTTPError.
func (e *Error) HTTPError() *echo.HTTPError {
	code := HTTPStatusCodeFromKind(e.kind)
	return echo.NewHTTPError(code, e.Error())
}

// New returns an error with the supplied kind and message. If message is empty, a default message
// for the error kind will be used.
func New(kind Kind, msg string) error {
	if msg == "" {
		msg = string(kind)
	}
	return &Error{
		error: errors.New(msg),
		kind:  kind,
	}
}

// Errorf formats according to a format specifier and return an unknown error with the string.
func Errorf(kind Kind, format string, args ...interface{}) error {
	return New(kind, fmt.Sprintf(format, args...))
}

// Wrap returns an error annotating err with a kind and a stacktrace at the point Wrap is called,
// and the supplied kind and message. If err is nil, Wrap returns nil.
func Wrap(err error, kind Kind, msg string) error {
	if err == nil {
		return nil
	}
	if msg == "" {
		msg = string(kind)
	}
	return &Error{
		error: errors.Wrap(err, msg),
		kind:  kind,
	}
}

// Wrapf returns an error annotating err with a stack trace at the point Wrapf is called, and the
// kind and format specifier. If err is nil, Wrapf returns nil.
func Wrapf(err error, kind Kind, format string, args ...interface{}) error {
	return Wrap(err, kind, fmt.Sprintf(format, args...))
}

// WithFieldViolations returns an error with supplied field
// violations.
func WithFieldViolations(kind Kind, msg string, fieldViolations []FieldViolation) error {
	if msg == "" {
		msg = string(kind)
	}
	return &Error{
		error:           errors.New(msg),
		kind:            kind,
		fieldViolations: fieldViolations,
	}
}

// WithValidateError maps a Validate error into an internal error representation.
func WithValidateError(err error) error {
	if err == nil {
		return nil
	}
	var fieldViolations []FieldViolation
	switch errWithType := err.(type) {
	// Error from validator
	case validator.ValidationErrors:
		for _, fieldError := range errWithType {
			fieldViolations = append(fieldViolations, FieldViolation{
				Field:       stringsUtil.ToUnderScore(fieldError.Field()),
				Description: fieldError.Tag(),
			})
		}
		// Assume any field violations corresponds to error returns from validator, which is invalid argument.
		return WithFieldViolations(ErrorInvalidArgument, "", fieldViolations)
	}
	return Wrap(err, ErrorUnknown, "")
}

// WithSQLError maps a SQL error into an internal error representation.
func WithSQLError(err error) error {
	if err == nil {
		return nil
	}
	if err == sql.ErrNoRows {
		return Wrap(err, ErrorNotFound, "")
	}
	switch err.(type) {
	case *mysql.MySQLError:
		sqlErr := err.(*mysql.MySQLError)
		switch sqlErr.Number {
		// Duplicated entry case
		case 1062:
			errItems := strings.Split(sqlErr.Message, " ")
			// Duplicated key will be the last item when the error message splits with space
			// Trim any single quote to get value
			errItem := strings.Trim(errItems[len(errItems)-1], "'")
			return WithFieldViolations(ErrorAlreadyExists, "", []FieldViolation{
				FieldViolation{
					Field:       errItem,
					Description: "duplicated",
				},
			})
		}
	}
	return Wrap(err, ErrorUnknown, "")
}

// HTTPStatusCodeFromKind converts an error kind into HTTP status code.
func HTTPStatusCodeFromKind(kind Kind) int {
	switch kind {
	case ErrorCanceled:
		return http.StatusRequestTimeout
	case ErrorUnknown:
		return http.StatusInternalServerError
	case ErrorInvalidArgument:
		return http.StatusBadRequest
	case ErrorDeadlineExceeded:
		return http.StatusGatewayTimeout
	case ErrorNotFound:
		return http.StatusNotFound
	case ErrorAlreadyExists:
		return http.StatusConflict
	case ErrorPermissionDenied:
		return http.StatusForbidden
	case ErrorUnauthenticated:
		return http.StatusUnauthorized
	case ErrorResourceExhausted:
		return http.StatusTooManyRequests
	case ErrorFailedPrecondition:
		// Note, this deliberately doesn't translate to the similarly named '412 Precondition Failed' HTTP response status.
		return http.StatusBadRequest
	case ErrorAborted:
		return http.StatusConflict
	case ErrorUnavailable:
		return http.StatusServiceUnavailable
	}

	log.Infof("Unknown error kind: %v", kind)
	return http.StatusInternalServerError
}

// GRPCStatusCodeFromKind converts an error kind into GRPC status code.
func GRPCStatusCodeFromKind(kind Kind) codes.Code {
	switch kind {
	case ErrorCanceled: // maps to HTTP 404
		return codes.Canceled
	case ErrorUnknown: // maps to HTTP 500
		return codes.Unknown
	case ErrorInvalidArgument: // map to HTTP 400
		return codes.InvalidArgument
	case ErrorDeadlineExceeded: // map to HTTP 504
		return codes.DeadlineExceeded
	case ErrorNotFound: // maps to HTTP 404
		return codes.NotFound
	case ErrorAlreadyExists: // map to HTTP 409
		return codes.AlreadyExists
	case ErrorPermissionDenied: // map to HTTP 403
		return codes.PermissionDenied
	case ErrorUnauthenticated: // map to HTTP 401
		return codes.Unauthenticated
	case ErrorResourceExhausted: // map to HTTP 429
		return codes.ResourceExhausted
	case ErrorFailedPrecondition: // map to HTTP 400
		return codes.FailedPrecondition
	case ErrorAborted: // map to HTTP 409
		return codes.Aborted
	case ErrorUnavailable: // map to HTTP 503
		return codes.Unavailable
	case ErrorUserTemporarilyBlocked: // map to HTTP 403
		return codes.PermissionDenied
	}
	log.Infof("Unknown error kind: %v", kind)
	return codes.Unknown
}

// IsKind checks whether any error in err's chain matches the error kind.
func IsKind(err error, kind Kind) bool {
	ie := &Error{}
	if As(err, &ie) {
		return ie.kind == kind
	}
	return false
}

// As finds the first error in err's chain that matches target, and if so, sets target to that
// error value and return true.
//
// Same as Go's errors.As
func As(err error, target interface{}) bool {
	return stderrors.As(err, target)
}

// Unwrap returns the result of calling the Unwrap method on err, if err's
// type contains an Unwrap method returning error.
// Otherwise, Unwrap returns nil.
//
// Same as Go's errors.Unwrap
func Unwrap(err error) error {
	return stderrors.Unwrap(err)
}
