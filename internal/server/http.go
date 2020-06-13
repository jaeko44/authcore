package server

import (
	"context"
	"net"

	"authcore.io/authcore/pkg/httputil"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

// HTTPAddr includes address field which should be filled using x-forwarded-for field from header
type HTTPAddr struct {
	Addr string
}

// Network function matches net.Addr interface, return null string matches with default value
func (addr HTTPAddr) Network() string {
	return ""
}

// String function matches net.Addr interface, return IP address with the port from peer
func (addr HTTPAddr) String() string {
	return addr.Addr
}

// HTTPAddressUnaryServerInterceptor changes the IP address in the peer context. Using the value from "x-forwarded-for" key in header when the address is 127.0.0.1, which indicates the request comes from HTTP/1.1.
func HTTPAddressUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		fromMD, ok := metadata.FromIncomingContext(ctx)
		var resp interface{}
		var err error
		if ok {
			fromPeerPtr, ok := peer.FromContext(ctx)
			if !ok {
				log.Fatalf("failed to get peer from context")
			}
			ipAddressResult := fromPeerPtr.Addr.String()
			address, _, err := net.SplitHostPort(fromPeerPtr.Addr.String())
			if err != nil {
				log.WithFields(log.Fields{
					"error": err,
				}).Info("cannot split host and port from address")
			}
			// For localhost address, refer to HTTP header
			if address == "127.0.0.1" {
				xForwardedFor := fromMD.Get("x-forwarded-for")
				if len(xForwardedFor) > 0 {
					ipAddressResult = httputil.GetIPAddrFromXFF(xForwardedFor[0])
				}
			}
			fromMD.Set("ip-address", ipAddressResult)
		} else {
			log.Fatalf("failed to get metadata from context")
		}
		if resp == nil {
			resp, err = handler(ctx, req)
		}
		return resp, err
	}
}
