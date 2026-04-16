package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"

	sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"
)

// TestDLQList_NoPermission_Returns403 verifies RBAC enforcement.
func TestDLQList_NoPermission_Returns403(t *testing.T) {
	dlq := NewDLQ(nil, nil)
	r := chi.NewRouter()

	// Auth middleware that sets an empty permission list.
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := sdamw.WithPermissions(r.Context(), []string{})
			ctx = sdamw.WithUserID(ctx, "test-user")
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})
	dlq.Routes(r)

	req := httptest.NewRequest(http.MethodGet, "/v1/admin/dlq", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}
