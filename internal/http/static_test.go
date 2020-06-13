package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestStatic(t *testing.T) {
	e := echo.New()
	e.GET("/api/v2/preferences", func(c echo.Context) error {
		c.String(http.StatusOK, "success")
		return nil
	})
	e.Group("/folder1", rewriteStaticMiddleware("../../test/static/folder1", 60, true))
	e.Group("/folder3", rewriteStaticMiddleware("../../test/static/folder3", 0, false))

	// No index
	req := httptest.NewRequest(http.MethodGet, "/folder1/folder2", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Equal(t, "", rec.Header().Get("Cache-Control"))

	// Has index
	req = httptest.NewRequest(http.MethodGet, "/folder1", nil)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "public, max-age=30", rec.Header().Get("Cache-Control"))
	assert.Equal(t, "\"success\"", rec.Body.String())

	// HTML5
	req = httptest.NewRequest(http.MethodGet, "/folder1/html5", nil)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "public, max-age=30", rec.Header().Get("Cache-Control"))
	assert.Equal(t, "text/html; charset=utf-8", rec.Header().Get("Content-Type"))
	assert.Equal(t, "\"success\"", rec.Body.String())

	// File found
	req = httptest.NewRequest(http.MethodGet, "/folder3/file1.json", nil)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "no-store", rec.Header().Get("Cache-Control"))
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	assert.Equal(t, "\"file1\"", rec.Body.String())

	// File not found
	req = httptest.NewRequest(http.MethodGet, "/folder3/none", nil)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Equal(t, "", rec.Header().Get("Cache-Control"))

	// Rewrite error
	req = httptest.NewRequest(http.MethodGet, "/folder3/file2.html", nil)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Equal(t, "", rec.Header().Get("Cache-Control"))
	assert.Equal(t, "{\"message\":\"Internal Server Error\"}\n", rec.Body.String())
}
