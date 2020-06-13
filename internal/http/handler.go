package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func rootHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.Redirect(http.StatusFound, "/web/")
	}
}

func healthzHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	}
}
