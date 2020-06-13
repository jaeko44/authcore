package http

import (
	"context"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

// GRPCGatewayRegisterFunc is a function pointer that registers the http handlers for a GRPC service.
type GRPCGatewayRegisterFunc func(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error

func grpcGatewayMiddleware(prefix string, registerFunc GRPCGatewayRegisterFunc) echo.MiddlewareFunc {
	grpcEndpoint := viper.GetString("grpc_listen")
	conn, err := grpc.Dial(grpcEndpoint, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect to grpc server: %v", err)
	}
	// For the marshaler options, `OrigName` and `EmitDefaults` are set:
	// - OrigName is used to maintain the original key name (which uses underscore_case instead of camelCase), and
	// - EmitDefaults is used to show the default values (they were originally hidden).
	// https://github.com/grpc-ecosystem/grpc-gateway/issues/233#issuecomment-253365396
	mux := runtime.NewServeMux(runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{OrigName: true, EmitDefaults: true}))
	ctx := context.Background()
	err = registerFunc(ctx, mux, conn)
	if err != nil {
		log.Fatalf("failed to register GRPC gateway: %v", err)
	}
	handler := http.StripPrefix(prefix, mux)
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return echo.WrapHandler(handler)
	}
}
