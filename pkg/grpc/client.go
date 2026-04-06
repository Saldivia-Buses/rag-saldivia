package sdagrpc

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
)

// Dial creates a gRPC client connection with standard SDA defaults.
// Uses insecure transport (internal Docker network only — never exposed externally).
// Connection is lazy — does not fail on unreachable target at dial time.
func Dial(target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	defaults := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                2 * time.Minute,
			Timeout:             20 * time.Second,
			PermitWithoutStream: true,
		}),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(DefaultMaxRecvMsgSize),
		),
	}
	return grpc.NewClient(target, append(defaults, opts...)...)
}

// ForwardJWT creates outgoing metadata with the JWT from the caller.
// Used by services that proxy requests to other services.
func ForwardJWT(ctx context.Context, jwt string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+jwt)
}
