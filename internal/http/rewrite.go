package http

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strings"

	"authcore.io/authcore/internal/errors"

	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

func rewriteResponseProxy(target string) echo.MiddlewareFunc {
	targetURL, err := url.Parse(target)
	if err != nil {
		log.Fatalf("invalid proxy target: %v", target)
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			res := c.Response()
			proxy := httputil.NewSingleHostReverseProxy(targetURL)
			proxy.ModifyResponse = rewriteResponse(c)
			proxy.ErrorHandler = func(resp http.ResponseWriter, req *http.Request, err error) {
				desc := targetURL.String()
				c.Set("_error", echo.NewHTTPError(http.StatusBadGateway, fmt.Sprintf("remote %s unreachable, could not forward: %v", desc, err)))
			}
			proxy.ServeHTTP(res, req)
			if e, ok := c.Get("_error").(error); ok {
				return e
			}
			return nil
		}
	}
}

func rewriteResponse(c echo.Context) func(*http.Response) error {
	return func(res *http.Response) error {
		if shouldRewrite(res.Header.Get("Content-Type")) {
			if res.StatusCode == http.StatusPartialContent {
				return errors.New(errors.ErrorUnknown, "partial content response is not supported")
			}
			// Rewrite the response body
			body, err := rewriteStream(c, res.Body)
			if err != nil {
				return err
			}
			// Disable caching for rewritten response to avoid partial cache.
			res.Header.Set("Cache-Control", "no-store")
			res.ContentLength = int64(body.Len())
			res.Body = ioutil.NopCloser(body)
		}

		// Return the original response when we can't process the body.
		return nil
	}
}

type rewriteFunc func(echo.Context, []byte) ([]byte, error)

var rewriteRegex = regexp.MustCompile(`<!--#(\w+).*?-->`)

var rewriteFuncMap = map[string]rewriteFunc{
	"authcore_settings": rewriteWithInternalRequest(http.MethodGet, "/api/v2/preferences"),
}

// rewriteStream reads the input Reader into memory, and return a Reader with the rewritten
// content.
func rewriteStream(c echo.Context, r io.ReadCloser) (*bytes.Reader, error) {
	defer r.Close()
	body, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("upstream error: %w", err)
	}

	body2 := rewriteRegex.ReplaceAllFunc(body, func(m []byte) []byte {
		g := rewriteRegex.FindSubmatch(m)
		directive := string(g[1])
		f, ok := rewriteFuncMap[directive]
		if ok {
			m2, e := f(c, m)
			if e != nil {
				log.Errorf("error occurred in rewrite response function %v: %v", directive, e)
				err = e
				return nil
			}
			return m2
		}
		err = fmt.Errorf("undefined rewrite response function %v", directive)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(body2), nil
}

func rewriteWithInternalRequest(method, target string) rewriteFunc {
	return func(c echo.Context, m []byte) ([]byte, error) {
		return internalRequest(c, method, target+"?"+c.QueryString(), nil)
	}
}

func internalRequest(c echo.Context, method, target string, body io.Reader) ([]byte, error) {
	// Constructing a new incoming server Request suitable for passing to a handler
	if method == "" {
		method = "GET"
	}
	req, err := http.ReadRequest(bufio.NewReader(strings.NewReader(method + " " + target + " HTTP/1.0\r\n\r\n")))
	if err != nil {
		log.Fatalf("invalid internal request arguments: %v", err)
	}

	// HTTP/1.0 was used above to avoid needing a Host field. Change it to 1.1 here.
	req.Proto = "HTTP/1.1"
	req.ProtoMinor = 1
	req.Close = false

	if body != nil {
		switch v := body.(type) {
		case *bytes.Buffer:
			req.ContentLength = int64(v.Len())
		case *bytes.Reader:
			req.ContentLength = int64(v.Len())
		case *strings.Reader:
			req.ContentLength = int64(v.Len())
		default:
			req.ContentLength = -1
		}
		if rc, ok := body.(io.ReadCloser); ok {
			req.Body = rc
		} else {
			req.Body = ioutil.NopCloser(body)
		}
	}

	req.RemoteAddr = c.Request().RemoteAddr

	req.Host = c.Request().Host

	req.TLS = &tls.ConnectionState{
		Version:           tls.VersionTLS12,
		HandshakeComplete: true,
		ServerName:        req.Host,
	}

	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := newRecorder()
	c.Echo().ServeHTTP(rec, req)
	bytes := rec.Body.Bytes()
	// Always encode the response as a Javascript string to prevent XSS.
	return json.Marshal(string(bytes))
}

func shouldRewrite(contenttype string) bool {
	mediatype, params, err := mime.ParseMediaType(contenttype)
	if err != nil {
		return false
	}
	if mediatype == "text/html" {
		charset := strings.ToLower(params["charset"])
		if charset == "" || charset == "utf-8" {
			return true
		}
	}
	return false
}
