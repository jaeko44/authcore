package xrequestid

import (
	"context"

	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type requestIDKey struct{}

// UnaryServerInterceptor receives request id from metadata and set the request id to context and grpc_ctxtags
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		requestID := FromContext(ctx)
		ctx = context.WithValue(ctx, requestIDKey{}, requestID)

		grpc.SetHeader(ctx, metadata.Pairs(XRequestIDKey, requestID))

		tags := grpc_ctxtags.Extract(ctx)
		tags.Set("xrequestid", requestID)

		return handler(ctx, req)
	}
}
