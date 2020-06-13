package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"authcore.io/authcore/internal/errors"

	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestServer(t *testing.T) {
	viper.Set("base_path", "../..")
	server := NewServer()

	req := httptest.NewRequest("GET", "/healthz", nil)
	rec := httptest.NewRecorder()
	server.e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestErrorHandler(t *testing.T) {
	viper.Set("base_path", "../..")
	server := NewServer()
	server.e.GET("/__test__/err", func(c echo.Context) error {
		err := errors.New(errors.ErrorDeadlineExceeded, "error message")
		return err
	})

	req := httptest.NewRequest("GET", "/__test__/err", nil)
	rec := httptest.NewRecorder()
	server.e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusGatewayTimeout, rec.Code)
	res := make(map[string]interface{})
	err := json.Unmarshal([]byte(rec.Body.String()), &res)
	assert.NoError(t, err)
	assert.Equal(t, "error message", res["message"])
}
