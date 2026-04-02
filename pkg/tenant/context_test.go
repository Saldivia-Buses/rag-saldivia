package tenant

import (
	"context"
	"testing"
)

func TestWithInfo_and_FromContext(t *testing.T) {
	ctx := context.Background()
	info := Info{ID: "t-123", Slug: "saldivia"}

	ctx = WithInfo(ctx, info)

	got, err := FromContext(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got.ID != info.ID || got.Slug != info.Slug {
		t.Fatalf("expected %+v, got %+v", info, got)
	}
}

func TestFromContext_NoTenant(t *testing.T) {
	ctx := context.Background()

	_, err := FromContext(ctx)
	if err != ErrNoTenant {
		t.Fatalf("expected ErrNoTenant, got %v", err)
	}
}

func TestSlugFromContext(t *testing.T) {
	ctx := WithInfo(context.Background(), Info{ID: "t-1", Slug: "empresa2"})

	slug := SlugFromContext(ctx)
	if slug != "empresa2" {
		t.Fatalf("expected 'empresa2', got %q", slug)
	}
}

func TestSlugFromContext_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic, got none")
		}
	}()

	SlugFromContext(context.Background())
}
