package sdagrpc

import (
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

// DefaultMaxRecvMsgSize is 4MB (sufficient for most RPCs).
const DefaultMaxRecvMsgSize = 4 * 1024 * 1024

// NewServer creates a gRPC server with standard SDA interceptors and limits.
// IMPORTANT: additional opts must use ChainUnaryInterceptor, NOT UnaryInterceptor
// (the latter would silently replace the auth interceptor chain).
func NewServer(cfg InterceptorConfig, opts ...grpc.ServerOption) *grpc.Server {
	defaults := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(DefaultMaxRecvMsgSize),
		grpc.ChainUnaryInterceptor(AuthUnaryInterceptor(cfg)),
		grpc.ChainStreamInterceptor(AuthStreamInterceptor(cfg)),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: 5 * time.Minute,
			MaxConnectionAge:  30 * time.Minute,
			Time:              2 * time.Minute,
			Timeout:           20 * time.Second,
		}),
	}
	return grpc.NewServer(append(defaults, opts...)...)
}
