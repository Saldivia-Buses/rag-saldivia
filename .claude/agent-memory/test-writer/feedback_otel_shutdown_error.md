---
name: OTel graceful degradation tests
description: Shutdown() returning error on unreachable OTLP endpoint is expected behavior — use t.Logf not t.Fatalf
type: feedback
---

OTel tests that call `Setup()` with an unreachable endpoint and then call `shutdown(ctx)` will receive a non-nil error. This is **correct graceful degradation behavior** — the exporter silently drops data when the collector is unreachable.

Using `t.Fatalf` on the shutdown error causes the test to fail. The fix is `t.Logf`.

Also: reduce the context timeout for Shutdown in tests from 5s to 2s to keep the test suite fast. The internal exporter timeout is 5s, so a 5s context just blocks until it expires.

**Why:** The OTLP gRPC exporter has an internal timeout (5s in this codebase). When the endpoint is unreachable, `mp.Shutdown(ctx)` blocks until context deadline and returns the flush error. Tests should verify no panic, not that the error is nil.

**How to apply:** In any test that calls `shutdown(ctx)` where the OTLP endpoint is fake/unreachable, use `t.Logf` for the error and set context timeout to 2s or less.
