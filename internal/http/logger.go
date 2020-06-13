package http

import (
	"fmt"
	"strconv"
	"time"

	"authcore.io/authcore/pkg/log"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

// Logger returns a middleware that logs HTTP requests using logrus logger.
func Logger() echo.MiddlewareFunc {

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			req := c.Request()
			res := c.Response()
			start := time.Now()

			// Set up a context logger to log with context info
			id := req.Header.Get(echo.HeaderXRequestID)
			if id == "" {
				id = res.Header().Get(echo.HeaderXRequestID)
			}
			contextLogger := logrus.WithFields(logrus.Fields{
				"id": id,
			})
			ctx := log.WithLogger(c.Request().Context(), contextLogger)
			c.SetRequest(c.Request().WithContext(ctx))

			if err = next(c); err != nil {
				c.Error(err)
			}
			stop := time.Now()
			fields := logrus.Fields{}
			fields["remote_ip"] = c.RealIP()
			fields["host"] = req.Host
			fields["method"] = req.Method
			fields["uri"] = req.RequestURI
			fields["user_agent"] = req.UserAgent()
			fields["status"] = res.Status
			if err != nil {
				fields["error"] = fmt.Sprintf("%+v", err)
			}
			l := stop.Sub(start)
			fields["latency"] = strconv.FormatInt(int64(l), 10)
			fields["latency_human"] = l.String()
			cl := req.Header.Get(echo.HeaderContentLength)
			if cl == "" {
				cl = "0"
			}
			fields["bytes_in"] = cl
			fields["bytes_out"] = res.Size
			user := c.Get("user_id")
			if user != nil {
				fields["user"] = user
			}

			lvl := logrus.InfoLevel
			n := res.Status
			switch {
			case n >= 500:
				lvl = logrus.ErrorLevel
			case n >= 400:
				lvl = logrus.WarnLevel
			}

			contextLogger.WithFields(fields).Log(lvl, "http request")

			return
		}
	}
}
