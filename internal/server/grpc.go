package server

import (
	"context"
	"fmt"

	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"google.golang.org/grpc"
)

// ErrorLoggingUnaryServerInterceptor is a grpc.UnaryServerInterceptor that logs errors.
func ErrorLoggingUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		resp, err := handler(ctx, req)
		if err == nil {
			return resp, nil
		}

		// Logs the stack trace for the errors as well
		tags := grpc_ctxtags.Extract(ctx)
		tags.Set("error_stack", fmt.Sprintf("%+v", err))
		return resp, err
	}
}
