package tools

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"
)

type fakeSearch struct {
	gotQuery string
	gotColl  string
	gotMax   int
	ret      any
	err      error
}

func (f *fakeSearch) SearchDocuments(_ context.Context, q, col string, max int) (any, error) {
	f.gotQuery, f.gotColl, f.gotMax = q, col, max
	return f.ret, f.err
}

type fakeIngest struct {
	gotUser  string
	gotLimit int
	ret      any
	err      error
}

func (f *fakeIngest) ListJobs(_ context.Context, userID string, limit int) (any, error) {
	f.gotUser, f.gotLimit = userID, limit
	return f.ret, f.err
}

func TestExecute_SearchBackend_Hit(t *testing.T) {
	t.Parallel()
	defs := []Definition{{Name: "search_documents", Service: "search", Type: "read"}}
	exec := NewExecutor(defs)

	sb := &fakeSearch{ret: map[string]any{"query": "q", "selections": []any{}}}
	exec.SetSearchBackend(sb)

	got, err := exec.Execute(context.Background(), "tok", "search_documents",
		json.RawMessage(`{"query":"q","collection_id":"c","max_nodes":3}`))
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if got.Status != "success" {
		t.Fatalf("status = %q, want success (err=%q)", got.Status, got.Error)
	}
	if sb.gotQuery != "q" || sb.gotColl != "c" || sb.gotMax != 3 {
		t.Fatalf("backend args = (%q,%q,%d)", sb.gotQuery, sb.gotColl, sb.gotMax)
	}
	// Data should be marshaled map from backend
	var out map[string]any
	if err := json.Unmarshal(got.Data, &out); err != nil {
		t.Fatalf("data unmarshal: %v (data=%s)", err, got.Data)
	}
}

func TestExecute_SearchBackend_Error(t *testing.T) {
	t.Parallel()
	defs := []Definition{{Name: "search_documents", Type: "read"}}
	exec := NewExecutor(defs)
	exec.SetSearchBackend(&fakeSearch{err: errors.New("boom")})

	got, _ := exec.Execute(context.Background(), "tok", "search_documents",
		json.RawMessage(`{"query":"q"}`))
	if got.Status != "error" {
		t.Fatalf("status = %q, want error", got.Status)
	}
}

func TestExecute_IngestBackend_Hit(t *testing.T) {
	t.Parallel()
	defs := []Definition{{Name: "check_job_status", Service: "ingest", Type: "read"}}
	exec := NewExecutor(defs)

	ib := &fakeIngest{ret: []any{map[string]any{"id": "job-1", "status": "ready"}}}
	exec.SetIngestBackend(ib)

	ctx := sdamw.WithUserID(context.Background(), "user-7")
	got, err := exec.Execute(ctx, "tok", "check_job_status", json.RawMessage(`{"limit":5}`))
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if got.Status != "success" {
		t.Fatalf("status = %q, want success (err=%q)", got.Status, got.Error)
	}
	if ib.gotUser != "user-7" || ib.gotLimit != 5 {
		t.Fatalf("backend args = (%q,%d)", ib.gotUser, ib.gotLimit)
	}
}

func TestExecute_IngestBackend_MissingIdentity(t *testing.T) {
	t.Parallel()
	defs := []Definition{{Name: "check_job_status", Type: "read"}}
	exec := NewExecutor(defs)
	exec.SetIngestBackend(&fakeIngest{})

	got, _ := exec.Execute(context.Background(), "tok", "check_job_status", nil)
	if got.Status != "denied" {
		t.Fatalf("status = %q, want denied when user id missing", got.Status)
	}
}
