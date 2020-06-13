package testutil

import (
	"context"
	"bytes"
	"encoding/json"
	"net/http/httptest"

	"github.com/labstack/echo/v4"

	"authcore.io/authcore/internal/errors"
)

type testKey struct{}

// JSONRequest makes a JSON request with the given Echo instance.
func JSONRequest(e *echo.Echo, method, path string, body map[string]interface{}, args ...interface{}) (int, map[string]interface{}, error) {
	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return -1, nil, err
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(bodyJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if len(args) >= 1 {
		c.Set("user", args[0])
	}
	if len(args) >= 2 {
		c.Set("session", args[1])
	}
	// In real server there is cancel context with new Done channel ready for cancelation without leaking the goroutine.
	// In sql the package checks the cancel context to return error if there is connection error. If any cases the connection
	// is corrupted (For example, rows are not scanned nor closed in transaction) without cancel context, no error is returned.
	//
	// To prevent missing cases for SQL error in test cases, providing cancel context to match with real server environment
	ctx := c.Request().Context()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	ctx = context.WithValue(ctx, testKey{}, "127.0.0.1")
	c.SetRequest(c.Request().WithContext(ctx))
	e.Router().Find(req.Method, req.URL.Path, c)
	err = c.Handler()(c)
	if err != nil {
		var code int
		if ie, ok := err.(*errors.Error); ok {
			code = errors.HTTPStatusCodeFromKind(ie.Kind())
		}
		return code, nil, err
	}
	if rec.Body.Len() != 0 {
		res := make(map[string]interface{})
		err = json.Unmarshal([]byte(rec.Body.String()), &res)
		return rec.Code, res, err
	}
	return rec.Code, nil, nil
}
