package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

// mockRedis implements RedisClient for testing.
type mockRedis struct {
	store  map[string]string
	getErr error
	setErr error
	delErr error
}

func newMockRedis() *mockRedis {
	return &mockRedis{store: make(map[string]string)}
}

func (m *mockRedis) Get(_ context.Context, key string) (string, error) {
	if m.getErr != nil {
		return "", m.getErr
	}
	v, ok := m.store[key]
	if !ok {
		return "", fmt.Errorf("key not found")
	}
	return v, nil
}

func (m *mockRedis) Set(_ context.Context, key string, value string, _ time.Duration) error {
	if m.setErr != nil {
		return m.setErr
	}
	m.store[key] = value
	return nil
}

func (m *mockRedis) Del(_ context.Context, keys ...string) error {
	if m.delErr != nil {
		return m.delErr
	}
	for _, k := range keys {
		delete(m.store, k)
	}
	return nil
}

func TestAvailable(t *testing.T) {
	tests := []struct {
		name   string
		client RedisClient
		want   bool
	}{
		{name: "nil client", client: nil, want: false},
		{name: "mock client", client: newMockRedis(), want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewJSONCache(tt.client)
			if got := c.Available(); got != tt.want {
				t.Errorf("Available() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNilClient_AllNoOps(t *testing.T) {
	c := NewJSONCache(nil)
	ctx := context.Background()

	// Get should return false, not panic
	var dest map[string]string
	if c.Get(ctx, "key", &dest) {
		t.Error("Get with nil client should return false")
	}

	// Set should not panic
	c.Set(ctx, "key", map[string]string{"a": "b"}, time.Minute)

	// Del should not panic
	c.Del(ctx, "key")
}

func TestSetAndGet(t *testing.T) {
	mock := newMockRedis()
	c := NewJSONCache(mock)
	ctx := context.Background()

	type item struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}

	original := item{Name: "test", Count: 42}
	c.Set(ctx, "item:1", original, 5*time.Minute)

	// Verify it's stored as JSON in the mock
	raw, exists := mock.store["item:1"]
	if !exists {
		t.Fatal("Set did not store value in mock")
	}

	var stored item
	if err := json.Unmarshal([]byte(raw), &stored); err != nil {
		t.Fatalf("stored value is not valid JSON: %v", err)
	}
	if stored != original {
		t.Errorf("stored = %+v, want %+v", stored, original)
	}

	// Get it back
	var got item
	if !c.Get(ctx, "item:1", &got) {
		t.Fatal("Get returned false for existing key")
	}
	if got != original {
		t.Errorf("Get returned %+v, want %+v", got, original)
	}
}

func TestGet_MissingKey(t *testing.T) {
	mock := newMockRedis()
	c := NewJSONCache(mock)
	ctx := context.Background()

	var dest string
	if c.Get(ctx, "nonexistent", &dest) {
		t.Error("Get should return false for missing key")
	}
}

func TestGet_InvalidJSON(t *testing.T) {
	mock := newMockRedis()
	mock.store["bad"] = "not-json{{"
	c := NewJSONCache(mock)
	ctx := context.Background()

	var dest map[string]string
	if c.Get(ctx, "bad", &dest) {
		t.Error("Get should return false for invalid JSON")
	}
}

func TestGet_EmptyString(t *testing.T) {
	mock := newMockRedis()
	mock.store["empty"] = ""
	c := NewJSONCache(mock)
	ctx := context.Background()

	var dest string
	if c.Get(ctx, "empty", &dest) {
		t.Error("Get should return false for empty string value")
	}
}

func TestGet_RedisError(t *testing.T) {
	mock := newMockRedis()
	mock.getErr = fmt.Errorf("connection refused")
	c := NewJSONCache(mock)
	ctx := context.Background()

	var dest string
	if c.Get(ctx, "key", &dest) {
		t.Error("Get should return false on Redis error")
	}
}

func TestDel(t *testing.T) {
	mock := newMockRedis()
	c := NewJSONCache(mock)
	ctx := context.Background()

	c.Set(ctx, "a", "val-a", time.Minute)
	c.Set(ctx, "b", "val-b", time.Minute)

	if len(mock.store) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(mock.store))
	}

	c.Del(ctx, "a", "b")

	if len(mock.store) != 0 {
		t.Errorf("expected 0 keys after Del, got %d", len(mock.store))
	}
}

func TestSet_MarshalError(t *testing.T) {
	mock := newMockRedis()
	c := NewJSONCache(mock)
	ctx := context.Background()

	// channels cannot be marshaled to JSON
	ch := make(chan int)
	c.Set(ctx, "bad", ch, time.Minute)

	if _, exists := mock.store["bad"]; exists {
		t.Error("Set should not store value when marshal fails")
	}
}

func TestSet_RedisError(t *testing.T) {
	mock := newMockRedis()
	mock.setErr = fmt.Errorf("write error")
	c := NewJSONCache(mock)
	ctx := context.Background()

	// Should not panic even if Redis returns error
	c.Set(ctx, "key", "value", time.Minute)
}
