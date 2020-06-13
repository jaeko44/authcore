package xrequestid

import (
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
)

// XRequestIDKey is metadata key name for request ID
var XRequestIDKey = "x-request-id"

// FromContext extracts request ID from grpc metadata.
func FromContext(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return newRequestID()
	}

	header, ok := md[XRequestIDKey]
	if !ok || len(header) == 0 {
		return newRequestID()
	}

	requestID := header[0]
	if requestID == "" {
		return newRequestID()
	}

	return fmt.Sprintf("%s,%s", requestID, newRequestID())
}

func newRequestID() string {
	return uuid.New().String()
}
