---
title: Package: pkg/plc
audience: ai
last_reviewed: 2026-04-15
related:
  - ../README.md
  - ./approval.md
  - ./remote.md
---

## Purpose

Industrial protocol clients (Modbus TCP and OPC-UA) plus a `SafetyTier`
classification used to gate writes. **WARNING:** raw protocol clients only —
production use against physical PLCs MUST go through a service that enforces
safety tiers, audit logging, rate limits, and credential management
(BigBrother is the canonical caller). Import this only when building such a
service or writing low-level tests.

## Public API

| Symbol | Source | Description |
|--------|--------|-------------|
| `SafetyTier` | `safety.go:17` | Enum: `unclassified`, `safe`, `controlled`, `critical` |
| `ValidTiers` | `safety.go:45` | Slice of all tier values |
| `SafetyTier.IsValid()` | `safety.go:48` | Recognized value? |
| `SafetyTier.AllowsWrite()` | `safety.go:59` | True for safe/controlled/critical |
| `SafetyTier.RequiresTwoPersonApproval()` | `safety.go:64` | True only for critical |
| `ValidateWrite(tier, value, min, max)` | `safety.go:74` | Range check; rejects unclassified |
| `ErrWriteNotAllowed`/`ErrValueOutOfRange`/`ErrUnclassified` | `safety.go:39` | Sentinel errors |
| `ModbusConfig` | `modbus.go:21` | `Address`, `Timeout` (default 10s) |
| `ModbusClient` | `modbus.go:13` | Mutex-protected wrapper, port 502 |
| `NewModbusClient(cfg)` | `modbus.go:28` | Constructor (does NOT connect) |
| `ModbusClient.Connect/Close` | | Open/close TCP |
| `ModbusClient.ReadHoldingRegisters / ReadInputRegisters / WriteRegister` | | Standard ops |
| `ModbusClient.ReadAndVerifyWrite` | `modbus.go:127` | Write then read-back verify |
| `Register` | `modbus.go:67` | `Address`, `Value`, `FloatVal` |
| `OPCUAConfig` | `opcua.go:23` | `Endpoint`, `SecurityMode`, certs |
| `OPCUAClient` | `opcua.go:16` | Mutex-protected wrapper |
| `NewOPCUAClient(cfg)` | `opcua.go:37` | Constructor (does NOT connect) |
| `OPCUAClient.Connect/Close/ReadNode/BrowseChildren/WriteNode` | | Standard ops |
| `OPCUANode` | `opcua.go:93` | `NodeID`, `Name`, `Value` |

## Usage

```go
// Always validate the tier before writing
if err := plc.ValidateWrite(tier, newValue, &cfg.Min, &cfg.Max); err != nil {
    return err
}
if tier.RequiresTwoPersonApproval() {
    // hand off to pkg/approval
}
c, _ := plc.NewModbusClient(plc.ModbusConfig{Address: addr})
_ = c.Connect(); defer c.Close()
err := c.ReadAndVerifyWrite(ctx, registerAddr, newValue)
```

## Invariants

- All clients are mutex-protected and safe for concurrent use, but a Modbus
  TCP connection serialises operations — high throughput needs multiple
  clients.
- `TierUnclassified` NEVER allows writes (`pkg/plc/safety.go:79`). Auto-
  discovered registers must be human-classified before they are writable.
- OPC-UA defaults to `SignAndEncrypt`. Setting `SecurityMode` to
  `MessageSecurityModeNone` logs a `slog.Warn` and is intended only for
  testing (`pkg/plc/opcua.go:43`).
- OPC-UA mutual TLS requires an RSA private key; non-RSA keys fail at
  construction (`pkg/plc/opcua.go:55`).
- This package does NOT publish audit events — that is the caller's
  responsibility (use `pkg/audit.StrictLogger`).

## Importers

`services/bigbrother/internal/service/plc.go` is the only production importer.
The package is intentionally narrow — it is not for direct use by other services.
