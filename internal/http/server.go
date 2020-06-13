package http

import (
	"path/filepath"

	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/internal/validator"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Server represents a HTTP server
type Server struct {
	e *echo.Echo
}

// NewServer returns a new Server instanece based on configuration. middlewares are additional
// middlewares to be added to the server, usually for authorization.
func NewServer(middlewares ...echo.MiddlewareFunc) *Server {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Validator = validator.Validator
	s := &Server{
		e: e,
	}
	s.initMiddleware(middlewares...)
	return s
}

// Start starts the Server
func (s *Server) Start() {
	httpsEnabled := viper.GetBool("https_enabled")
	if httpsEnabled {
		go s.startHTTPS()
	}
	address := viper.GetString("http_listen")
	log.Infof("http server started on %v", address)
	log.Fatal(s.e.Start(address))
}

// GRPCGateway registers HTTP handlers for a GRPC gateway service.
func (s *Server) GRPCGateway(prefix string, registerFunc GRPCGatewayRegisterFunc) {
	s.e.Group(prefix, grpcGatewayMiddleware(prefix, registerFunc))
}

// Register registers new HTTP routes using the given RegisterFunc.
func (s *Server) Register(registerFunc RegisterFunc) {
	registerFunc(s.e)
}

// Echo returns the echo instance.
func (s *Server) Echo() *echo.Echo {
	return s.e
}

// httpErrorHandler is an HTTP error handler that converts internal errors into HTTPError and
// handle it using echo.DefaultHTTPErrorHandler.
func (s *Server) httpErrorHandler(err error, c echo.Context) {
	if ie, ok := err.(*errors.Error); ok {
		err = ie.HTTPError()
	}
	s.e.DefaultHTTPErrorHandler(err, c)
}

func (s *Server) initMiddleware(middlewares ...echo.MiddlewareFunc) {
	basePath := viper.GetString("base_path")
	staticPath := filepath.Join(basePath, "/web/static")
	webPath := filepath.Join(basePath, "/web/dist/web")
	widgetsPath := filepath.Join(basePath, "/web/dist/widgets")
	docsPath := filepath.Join(basePath, "/web/dist/docs")
	apiPath := filepath.Join(basePath, "/api")
	staticCacheTTL := viper.GetInt("static_cache_ttl")

	e := s.e
	e.HTTPErrorHandler = s.httpErrorHandler
	e.Use(middleware.Recover())
	e.Use(middleware.Gzip())
	e.Use(middleware.RequestID())
	e.Use(middleware.CORS())
	e.Use(Logger())
	e.Use(middleware.Static(staticPath))
	e.Use(middlewares...)
	e.GET("/", rootHandler())
	e.GET("/healthz", healthzHandler())
	e.Static("/api", apiPath)
	if viper.IsSet("web_url") {
		e.Group("/web", rewriteResponseProxy(viper.GetString("web_url")))
	} else {
		e.Group("/web", rewriteStaticMiddleware(webPath, staticCacheTTL, true))
	}
	if viper.IsSet("widgets_url") {
		e.Group("/widgets", rewriteResponseProxy(viper.GetString("widgets_url")))
	} else {
		e.Group("/widgets", rewriteStaticMiddleware(widgetsPath, staticCacheTTL, true))
	}
	if viper.GetBool("docs_enabled") {
		e.Static("/docs", docsPath)
	}
}

func (s *Server) startHTTPS() {
	address := viper.GetString("https_listen")
	cert := viper.GetString("https_cert")
	key := viper.GetString("https_key")
	log.Infof("https server started on %v", address)
	log.Fatal(s.e.StartTLS(address, cert, key))
}

// RegisterFunc registers new HTTP routes with the given Echo instance.
type RegisterFunc func(*echo.Echo)
