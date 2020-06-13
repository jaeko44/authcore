package http

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestRewriteResponseProxy(t *testing.T) {
	t1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "<!--#authcore_settings     -->")
	}))
	t2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "<!--#authcore_settings     -->")
	}))
	t3 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "<!--#undefined -->")
	}))

	e := echo.New()
	e.GET("/api/v2/preferences", func(c echo.Context) error {
		c.String(http.StatusOK, "success")
		return nil
	})
	e.Group("/t1", rewriteResponseProxy(t1.URL))
	e.Group("/t2", rewriteResponseProxy(t2.URL))
	e.Group("/t3", rewriteResponseProxy(t3.URL))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/t1", nil)
	e.ServeHTTP(rec, req)
	body := rec.Body.String()
	assert.Equal(t, "\"success\"", body)

	// No rewrite if content type is not "text/html"
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/t2", nil)
	e.ServeHTTP(rec, req)
	body = rec.Body.String()
	assert.Equal(t, "<!--#authcore_settings     -->", body)

	// Undefined rewrite function
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/t3", nil)
	e.ServeHTTP(rec, req)
	body = rec.Body.String()
	assert.Equal(t, http.StatusBadGateway, rec.Code)
	assert.Contains(t, body, "undefined rewrite response function undefined")
}
