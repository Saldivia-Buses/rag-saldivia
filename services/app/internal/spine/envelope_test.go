package spine_test

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/Camionerou/rag-saldivia/services/app/internal/spine"
)

type fakePayload struct {
	UserID string `json:"user_id"`
	Msg    string `json:"msg"`
}

func TestNew_SetsDefaults(t *testing.T) {
	p := fakePayload{UserID: "u1", Msg: "hi"}
	env, err := spine.New("saldivia", "chat.new_message", 1, p)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if env.ID == uuid.Nil {
		t.Error("expected non-nil ID")
	}
	if env.TenantID != "saldivia" {
		t.Errorf("TenantID = %q, want saldivia", env.TenantID)
	}
	if env.Type != "chat.new_message" {
		t.Errorf("Type = %q", env.Type)
	}
	if env.SchemaVersion != 1 {
		t.Errorf("SchemaVersion = %d", env.SchemaVersion)
	}
	if env.OccurredAt.IsZero() || env.RecordedAt.IsZero() {
		t.Error("timestamps not populated")
	}
	if env.Payload != p {
		t.Errorf("Payload = %+v, want %+v", env.Payload, p)
	}
}

func TestNew_GeneratesUUIDv7(t *testing.T) {
	env1, _ := spine.New("t", "x", 1, fakePayload{})
	env2, _ := spine.New("t", "x", 1, fakePayload{})
	if env1.ID == env2.ID {
		t.Error("two New calls produced the same UUID")
	}
	// UUID v7 has version bits 0111 in byte 6 upper nibble.
	if (env1.ID[6] >> 4) != 7 {
		t.Errorf("expected UUIDv7, got version %d", env1.ID[6]>>4)
	}
}

func TestNew_Validation(t *testing.T) {
	tests := []struct {
		name      string
		tenantID  string
		eventType string
		version   uint8
	}{
		{"empty tenant", "", "chat.new_message", 1},
		{"empty type", "saldivia", "", 1},
		{"zero version", "saldivia", "chat.new_message", 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := spine.New(tc.tenantID, tc.eventType, tc.version, fakePayload{}); err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestRoundTrip_EncodeDecode(t *testing.T) {
	env, err := spine.New("saldivia", "chat.new_message", 1, fakePayload{UserID: "u1", Msg: "hi"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	// Add optional fields to test round-trip of omitempty values.
	cid := uuid.New()
	env.CausationID = &cid
	env.CorrelationID = "corr-123"
	actor := "admin@saldivia"
	env.ActorUserID = &actor

	data, err := spine.Encode(env)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	decoded, err := spine.Decode[fakePayload](data)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if decoded.ID != env.ID {
		t.Errorf("ID: got %s want %s", decoded.ID, env.ID)
	}
	if decoded.Payload != env.Payload {
		t.Errorf("Payload: got %+v want %+v", decoded.Payload, env.Payload)
	}
	if decoded.CausationID == nil || *decoded.CausationID != cid {
		t.Errorf("CausationID roundtrip failed: %v", decoded.CausationID)
	}
	if decoded.CorrelationID != "corr-123" {
		t.Errorf("CorrelationID: got %q", decoded.CorrelationID)
	}
	if decoded.ActorUserID == nil || *decoded.ActorUserID != actor {
		t.Errorf("ActorUserID roundtrip failed: %v", decoded.ActorUserID)
	}
}

func TestDecode_RejectsIncomplete(t *testing.T) {
	validID := uuid.Must(uuid.NewV7()).String()
	now := time.Now().UTC().Format(time.RFC3339Nano)

	build := func(fields map[string]any) []byte {
		b, _ := json.Marshal(fields)
		return b
	}

	tests := []struct {
		name   string
		fields map[string]any
	}{
		{"missing id", map[string]any{
			"tenant_id": "t", "type": "x", "schema_version": 1,
			"occurred_at": now, "recorded_at": now, "payload": fakePayload{},
		}},
		{"missing tenant_id", map[string]any{
			"id": validID, "type": "x", "schema_version": 1,
			"occurred_at": now, "recorded_at": now, "payload": fakePayload{},
		}},
		{"missing type", map[string]any{
			"id": validID, "tenant_id": "t", "schema_version": 1,
			"occurred_at": now, "recorded_at": now, "payload": fakePayload{},
		}},
		{"zero schema_version", map[string]any{
			"id": validID, "tenant_id": "t", "type": "x", "schema_version": 0,
			"occurred_at": now, "recorded_at": now, "payload": fakePayload{},
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := spine.Decode[fakePayload](build(tc.fields)); err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestDecode_RejectsInvalidJSON(t *testing.T) {
	if _, err := spine.Decode[fakePayload]([]byte("not json")); err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestPeekHeader_Works(t *testing.T) {
	env, _ := spine.New("saldivia", "chat.new_message", 1, fakePayload{UserID: "u1"})
	data, _ := spine.Encode(env)

	h, err := spine.PeekHeader(data)
	if err != nil {
		t.Fatalf("PeekHeader: %v", err)
	}
	if h.ID != env.ID {
		t.Errorf("ID: got %s want %s", h.ID, env.ID)
	}
	if h.Type != "chat.new_message" {
		t.Errorf("Type: got %q", h.Type)
	}
	if h.SchemaVersion != 1 {
		t.Errorf("SchemaVersion: got %d", h.SchemaVersion)
	}
}

func TestPeekHeader_RejectsIncomplete(t *testing.T) {
	if _, err := spine.PeekHeader([]byte(`{}`)); err == nil {
		t.Error("expected error for empty object")
	}
}

func TestCheckSchemaVersion(t *testing.T) {
	env, _ := spine.New("t", "chat.new_message", 2, fakePayload{})

	if err := spine.CheckSchemaVersion(env, 2); err != nil {
		t.Errorf("matching version should be OK, got %v", err)
	}

	err := spine.CheckSchemaVersion(env, 1)
	if err == nil {
		t.Fatal("expected mismatch error")
	}
	var sv *spine.ErrSchemaVersion
	if !errors.As(err, &sv) {
		t.Fatalf("expected *ErrSchemaVersion, got %T", err)
	}
	if sv.Type != "chat.new_message" {
		t.Errorf("Type: got %q", sv.Type)
	}
	if sv.Expected != 1 || sv.Got != 2 {
		t.Errorf("Expected=%d Got=%d", sv.Expected, sv.Got)
	}
}

func TestEncode_IsValidJSON(t *testing.T) {
	env, _ := spine.New("t", "x", 1, fakePayload{UserID: "u", Msg: "m"})
	data, err := spine.Encode(env)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	var generic map[string]any
	if err := json.Unmarshal(data, &generic); err != nil {
		t.Fatalf("Encoded output is not valid JSON: %v", err)
	}
	for _, key := range []string{"id", "tenant_id", "type", "schema_version", "occurred_at", "recorded_at", "payload"} {
		if _, ok := generic[key]; !ok {
			t.Errorf("missing JSON field %q", key)
		}
	}
}
