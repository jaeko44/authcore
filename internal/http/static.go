package http

import (
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// rewriteStaticMiddleware returns a middleware to serve static files. If the requested file is a
// html file, it will be passed through rewriteStream to rewrite the content.
func rewriteStaticMiddleware(root string, cacheTTL int, html5 bool) echo.MiddlewareFunc {
	static := middleware.StaticWithConfig(middleware.StaticConfig{
		Root:  root,
		HTML5: html5,
	})
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		next = static(next)
		return func(c echo.Context) (err error) {
			return next(&rewriteFileWrapper{Context: c, cacheTTL: cacheTTL})
		}
	}
}

// rewriteContextWrapper wraps a echo.Context and override the File method to rewrite file content.
type rewriteFileWrapper struct {
	echo.Context
	cacheTTL int
}

func (c *rewriteFileWrapper) File(name string) (err error) {
	f, err := os.Open(name)
	if err != nil {
		return echo.NotFoundHandler(c)
	}
	defer f.Close()

	fi, _ := f.Stat()
	if fi.IsDir() {
		return echo.ErrForbidden
	}

	ext := filepath.Ext(fi.Name())
	mediatype := mime.TypeByExtension(ext)
	var r io.ReadSeeker
	if shouldRewrite(mediatype) {
		r, err = rewriteStream(c, f)
		if err != nil {
			return
		}
		if c.cacheTTL > 0 {
			// Allow a short TTL for dynamic content
			c.Response().Header().Set("Cache-Control", "public, max-age=30")
		}
	} else {
		if c.cacheTTL > 0 {
			c.Response().Header().Set("Cache-Control", "public, max-age="+strconv.Itoa(c.cacheTTL))
		}
		r = f
	}
	if c.cacheTTL <= 0 {
		c.Response().Header().Set("Cache-Control", "no-store")
	}
	http.ServeContent(c.Response(), c.Request(), fi.Name(), fi.ModTime(), r)
	return
}
