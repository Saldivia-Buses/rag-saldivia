// Package sdagrpc provides shared gRPC server/client factories and auth
// interceptors for SDA Framework inter-service communication.
//
// The auth interceptor verifies JWT from gRPC metadata with full parity
// to the HTTP middleware (pkg/middleware/auth.go): tenant context, role,
// permissions, blacklist check, and MFA-pending rejection.
package sdagrpc

import (
	"context"
	"crypto/ed25519"
	"log/slog"
	"strings"

	sdajwt "github.com/Camionerou/rag-saldivia/pkg/jwt"
	sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"
	"github.com/Camionerou/rag-saldivia/pkg/security"
	"github.com/Camionerou/rag-saldivia/pkg/tenant"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// InterceptorConfig holds dependencies for the gRPC auth interceptor.
type InterceptorConfig struct {
	PublicKey ed25519.PublicKey
	Blacklist *security.TokenBlacklist // nil = no blacklist
	FailOpen  bool                     // true = allow on Redis error, false = reject
}

// AuthUnaryInterceptor verifies JWT from gRPC metadata and injects
// tenant context, role, and permissions. Full parity with HTTP middleware.
func AuthUnaryInterceptor(cfg InterceptorConfig) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if info.FullMethod == "/grpc.health.v1.Health/Check" {
			return handler(ctx, req)
		}
		newCtx, err := extractAndVerifyJWT(ctx, cfg)
		if err != nil {
			return nil, err
		}
		return handler(newCtx, req)
	}
}

// AuthStreamInterceptor is the streaming equivalent.
func AuthStreamInterceptor(cfg InterceptorConfig) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if info.FullMethod == "/grpc.health.v1.Health/Check" {
			return handler(srv, ss)
		}
		newCtx, err := extractAndVerifyJWT(ss.Context(), cfg)
		if err != nil {
			return err
		}
		return handler(srv, &wrappedStream{ss, newCtx})
	}
}

func extractAndVerifyJWT(ctx context.Context, cfg InterceptorConfig) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing metadata")
	}
	vals := md.Get("authorization")
	if len(vals) == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing authorization")
	}
	token := vals[0]
	if strings.HasPrefix(token, "Bearer ") {
		token = token[7:]
	}

	claims, err := sdajwt.Verify(cfg.PublicKey, token)
	if err != nil {
		slog.Warn("grpc auth failed", "error", err)
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	// Blacklist check (parity with HTTP middleware)
	if cfg.Blacklist != nil {
		if claims.ID == "" {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}
		revoked, err := cfg.Blacklist.IsRevoked(ctx, claims.ID)
		if err != nil {
			slog.Error("grpc blacklist check failed", "error", err)
			if !cfg.FailOpen {
				return nil, status.Error(codes.Unavailable, "auth check unavailable")
			}
		} else if revoked {
			return nil, status.Error(codes.Unauthenticated, "token revoked")
		}
	}

	// Reject MFA-pending tokens
	if claims.Role == "mfa_pending" {
		return nil, status.Error(codes.Unauthenticated, "mfa verification required")
	}

	// Inject tenant, role, permissions into context (full parity with HTTP)
	ctx = tenant.WithInfo(ctx, tenant.Info{
		ID:   claims.TenantID,
		Slug: claims.Slug,
	})
	ctx = sdamw.WithRole(ctx, claims.Role)
	ctx = sdamw.WithPermissions(ctx, claims.Permissions)

	return ctx, nil
}

// JWTFromIncomingContext extracts the bearer token from incoming gRPC metadata.
// Used for gRPC-to-gRPC JWT forwarding.
func JWTFromIncomingContext(ctx context.Context) (string, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", false
	}
	vals := md.Get("authorization")
	if len(vals) == 0 {
		return "", false
	}
	token := vals[0]
	if strings.HasPrefix(token, "Bearer ") {
		return token[7:], true
	}
	return token, true
}

// wrappedStream overrides Context() for stream interceptors.
type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedStream) Context() context.Context { return w.ctx }
