package sdagrpc

import (
	"context"
	"crypto/ed25519"
	"testing"

	sdajwt "github.com/Camionerou/rag-saldivia/pkg/jwt"
	"google.golang.org/grpc/metadata"
)

func testKeys(t *testing.T) (ed25519.PrivateKey, ed25519.PublicKey) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	return priv, pub
}

func TestExtractAndVerifyJWT_ValidToken(t *testing.T) {
	priv, pub := testKeys(t)

	cfg := sdajwt.DefaultConfig(priv, pub)
	token, err := sdajwt.CreateAccess(cfg, sdajwt.Claims{
		UserID:   "u-1",
		Email:    "test@sda.app",
		TenantID: "t-1",
		Slug:     "test",
		Role:     "admin",
	})
	if err != nil {
		t.Fatal(err)
	}

	md := metadata.New(map[string]string{"authorization": "Bearer " + token})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	newCtx, err := extractAndVerifyJWT(ctx, InterceptorConfig{PublicKey: pub})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if newCtx == nil {
		t.Fatal("expected non-nil context")
	}
}

func TestExtractAndVerifyJWT_MissingMetadata(t *testing.T) {
	_, pub := testKeys(t)
	_, err := extractAndVerifyJWT(context.Background(), InterceptorConfig{PublicKey: pub})
	if err == nil {
		t.Fatal("expected error for missing metadata")
	}
}

func TestExtractAndVerifyJWT_InvalidToken(t *testing.T) {
	_, pub := testKeys(t)
	md := metadata.New(map[string]string{"authorization": "Bearer invalid.token.here"})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	_, err := extractAndVerifyJWT(ctx, InterceptorConfig{PublicKey: pub})
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}

func TestExtractAndVerifyJWT_MFAPending_Rejected(t *testing.T) {
	priv, pub := testKeys(t)

	cfg := sdajwt.DefaultConfig(priv, pub)
	token, _ := sdajwt.CreateAccess(cfg, sdajwt.Claims{
		UserID:   "u-1",
		TenantID: "t-1",
		Slug:     "test",
		Role:     "mfa_pending",
	})

	md := metadata.New(map[string]string{"authorization": "Bearer " + token})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	_, err := extractAndVerifyJWT(ctx, InterceptorConfig{PublicKey: pub})
	if err == nil {
		t.Fatal("expected error for mfa_pending token")
	}
}

func TestExtractAndVerifyJWT_WrongKey_Rejected(t *testing.T) {
	priv, _ := testKeys(t)
	_, otherPub := testKeys(t) // different key pair

	cfg := sdajwt.DefaultConfig(priv, nil)
	cfg.PublicKey = nil // only need private for signing
	token, _ := sdajwt.CreateAccess(sdajwt.DefaultConfig(priv, otherPub), sdajwt.Claims{
		UserID: "u-1", TenantID: "t-1", Slug: "test", Role: "user",
	})

	md := metadata.New(map[string]string{"authorization": "Bearer " + token})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	_, err := extractAndVerifyJWT(ctx, InterceptorConfig{PublicKey: otherPub})
	if err == nil {
		t.Fatal("expected error for wrong signing key")
	}
}

func TestForwardJWT_RoundTrip(t *testing.T) {
	ctx := ForwardJWT(context.Background(), "test-token-123")

	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		t.Fatal("expected outgoing metadata")
	}
	vals := md.Get("authorization")
	if len(vals) != 1 || vals[0] != "Bearer test-token-123" {
		t.Errorf("expected 'Bearer test-token-123', got %v", vals)
	}
}

func TestJWTFromIncomingContext_Present(t *testing.T) {
	md := metadata.New(map[string]string{"authorization": "Bearer my-jwt"})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	token, ok := JWTFromIncomingContext(ctx)
	if !ok || token != "my-jwt" {
		t.Errorf("expected 'my-jwt', got %q (ok=%v)", token, ok)
	}
}

func TestJWTFromIncomingContext_Missing(t *testing.T) {
	_, ok := JWTFromIncomingContext(context.Background())
	if ok {
		t.Fatal("expected not ok for missing metadata")
	}
}

func TestNewServer_NoPanic(t *testing.T) {
	_, pub := testKeys(t)
	srv := NewServer(InterceptorConfig{PublicKey: pub})
	if srv == nil {
		t.Fatal("expected non-nil server")
	}
	srv.Stop()
}

func TestDial_ReturnsConnection(t *testing.T) {
	conn, err := Dial("localhost:0")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if conn == nil {
		t.Fatal("expected non-nil connection")
	}
	conn.Close()
}
