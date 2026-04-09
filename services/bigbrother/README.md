# BigBrother — Network Intelligence Service

Network discovery, fingerprinting, PLC control, and remote execution
for the enterprise LAN. Discovers ~500 devices, classifies them, and
exposes them via REST API and agent tools.

## What it does

- **ARP discovery** — detects devices on the physical LAN
- **Fingerprinting** — nmap port scan, SNMP walk, mDNS, OS detection
- **PLC control** — Modbus TCP + OPC-UA read/write with safety tiers
- **Remote exec** — SSH (Linux) + WinRM HTTPS (Windows), enum-based commands
- **Credential vault** — envelope encryption (KEK/DEK + AAD)
- **Auto-documentation** — generates device tech sheets

## Endpoints

| Method | Path | Permission | Description |
|---|---|---|---|
| GET | `/v1/bigbrother/devices` | `bigbrother.read` | List devices |
| GET | `/v1/bigbrother/devices/{id}` | `bigbrother.read` | Device detail |
| GET | `/v1/bigbrother/devices/{id}/registers` | `bigbrother.plc.read` | PLC registers |
| POST | `/v1/bigbrother/devices/{id}/exec` | `bigbrother.exec` | Remote exec |
| POST | `/v1/bigbrother/devices/{id}/registers/{addr}` | `bigbrother.plc.write` | Write PLC |
| POST | `/v1/bigbrother/devices/{id}/registers/{addr}/approve` | `bigbrother.admin` | Approve critical |
| GET | `/v1/bigbrother/events` | `bigbrother.read` | Event timeline |
| GET | `/v1/bigbrother/topology` | `bigbrother.read` | Network map |
| POST | `/v1/bigbrother/scan` | `bigbrother.admin` | Trigger scan |
| POST | `/v1/bigbrother/scan/mode` | `bigbrother.admin` | Change scan mode |
| GET | `/v1/bigbrother/stats` | `bigbrother.read` | Network stats |
| POST | `/v1/bigbrother/credentials` | `bigbrother.admin` | Store credential |
| GET | `/v1/bigbrother/credentials` | `bigbrother.admin` | List credentials |
| DELETE | `/v1/bigbrother/credentials/{id}` | `bigbrother.admin` | Delete credential |

## NATS events

| Produces | Subject |
|---|---|
| Device discovered | `tenant.{slug}.bigbrother.device.discovered` |
| Device offline | `tenant.{slug}.bigbrother.device.offline` |
| Device online | `tenant.{slug}.bigbrother.device.online` |
| PLC value changed | `tenant.{slug}.bigbrother.plc.value_changed` |
| Approval requested | `tenant.{slug}.bigbrother.plc.approval_requested` |
| Scan completed | `tenant.{slug}.bigbrother.scan.completed` |

## Security profile

BigBrother has an **exceptional** security profile vs other SDA services:
- Alpine container (needs nmap binary)
- CAP_NET_RAW (ARP scanning)
- 4 Docker networks (frontend, backend, data, lan)
- Seccomp allowlist profile

See `docs/plans/2.0.x-plan15-bigbrother.md` for full security model.

## Development

```bash
cd services/bigbrother
go run ./cmd/bigbrother/
```

Note: ARP scanning requires a physical NIC (not available in WSL2).
Use `NetworkScanner` interface with mock/stub for local development.
