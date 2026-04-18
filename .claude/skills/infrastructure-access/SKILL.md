---
name: infrastructure-access
description: Use when the task needs to reach any Saldivia internal host — the SDA workstation, the legacy Histrix ERP, NAS, Proxmox hosts, VoIP, domain controller, or any node on the 172.22.0.0/16 corporate network. Covers the inventory of servers (names, IPs, roles, access method), the WireGuard VPN requirement for off-site access, and the iron rule — credentials are never committed, never logged, always requested from the user in the moment.
---

# infrastructure-access

Scope: any action that connects to a host on Saldivia's internal network. SSH,
RDP, HTTPS admin panels, DB consoles, file shares — all of it.

## The iron rule

**Credentials live with the user, not in the repo.**

- No password, SSH key, API token, WireGuard PrivateKey, RDP secret, or
  service-account credential is ever committed. Not in `.env`, not in a comment,
  not in a doc, not in a script.
- When Claude needs a credential to proceed, Claude **asks the user for it in
  the moment** via `AskUserQuestion`. The user pastes it from their password
  manager; Claude uses it for the current command; the value is not persisted.
- If Claude discovers a credential already committed, it is a **blocking finding**.
  Report it, recommend rotation, do not silently keep working.

## Network layout

Saldivia's internal network is the `172.22.0.0/16` block, carved into subnets:

| Subnet | What lives there |
|---|---|
| `172.22.52.0/24` | Firewall, DNS (`172.22.52.1`) |
| `172.22.60.0/24` | Domain-joined PCs |
| `172.22.68.0/24` | Mikrotik routers, PLCs |
| `172.22.70.0/24` | VoIP (FreePBX) |
| `172.22.80.0/24` | NVR / cameras |
| `172.22.90.0/24` | Ubiquiti wifi / nanostations |
| `172.22.100.0/24` | **Core servers + VMs (most work lives here)** |

A few hosts live in `192.168.1.0/24` / `192.168.2.0/24` (physical sala generadores).

Internal DNS: `172.22.52.1`.

## VPN (required from outside the office)

Off-site access goes through a WireGuard tunnel to the firewall. Every dev has
their own peer config with a personal PrivateKey — **the config file lives on
the dev machine, never in the repo**.

On macOS: use the WireGuard app and import a `.conf` the user keeps in their
password manager. On Linux: `/etc/wireguard/<name>.conf` with `0600` perms.

What a valid config contains (without pasting secrets here):

- `[Interface]` — the dev's PrivateKey + assigned Address + DNS `172.22.52.1`.
- `[Peer]` — the firewall's PublicKey, `AllowedIPs` for the corporate subnets
  above, the public `Endpoint`, and `PersistentKeepalive = 25`.

If the VPN is up, the internal subnets are reachable. If it isn't, nothing
inside `172.22.0.0/16` resolves.

### If the VPN browser route misbehaves

Occasionally the VPN routes traffic but the browser can't load an internal
dashboard. Fall back to SSH local-forward:

```bash
ssh -L <local-port>:<internal-host>:<remote-port> <user>@<jump-host>
# e.g. reach Zentyal (172.22.100.108:8443) through Histrix:
ssh -L 8443:172.22.100.108:8443 sistemas@172.22.100.99
# then open https://localhost:8443
```

## Server inventory

Core hosts. **Usernames** listed for orientation; **passwords are never here** —
ask the user in the moment via `AskUserQuestion`.

### SDA / RAG

| Host | IP | User | Role | Access |
|---|---|---|---|---|
| `srv-ia-01` (workstation) | `172.22.100.23` | `sistemas` | SDA backend + GPU — all Go services, SGLang, Postgres, NATS, Redis, MinIO, Traefik | SSH |

**Workstation specs** (see `deploy-ops` for implications):
AMD Threadripper Pro 9975WX (32c/64t) · 256 GB DDR5 ECC · 8 TB M.2 · 1× RTX PRO 6000 Blackwell Q Edition (96 GB VRAM).

### Legacy ERP + apps

| Host | IP | User | Role | Access |
|---|---|---|---|---|
| `Histrix` | `172.22.100.99` | `sistemas` | **Legacy ERP** — the source system RAG ingests from | SSH |
| `Vodemia` | `172.22.100.100` | `tecnico` | Legacy ERP companion app (VM on VServer02) | Console |
| `SRVRadan` | `172.22.100.203` | `sistemas` | Radan + SolidWorks apps | RDP |
| `Softcrates/Relojes` | `172.22.100.101` | `Administrador` | Attendance clocks app | RDP |
| `Calipso` | `172.22.100.110` | `Administrador` | Calipso application | RDP |

### Infra / servers

| Host | IP | User | Role | Access |
|---|---|---|---|---|
| `VServer02` | `172.22.100.2` | `root` | Proxmox VM host | SSH |
| `VServer05` | `172.22.100.5` | `root` | Proxmox VM host | SSH |
| `FServer01` (tecpia) | `172.22.100.21` | `root` | TrueNAS file server | SSH / web |
| `BKServer01` | `172.22.100.20` | `root` | TrueNAS backup server | SSH / web |
| `Ignacio` | `192.168.1.105` | `root` | Physical host, separate subnet | SSH |

### Network / management

| Host | IP | User | Role | Access |
|---|---|---|---|---|
| `Firewall` (Untangle) | `172.22.52.1:8080` | `admin` | Perimeter firewall + VPN endpoint | Web |
| `Zentyal` | `172.22.100.108:8443` | `admin` | AD / domain controller | Web |
| `UnifiController` / `glpi` | `172.22.100.98` | `root` | Ubiquiti + GLPI | Web |
| `Mikrotik Celulares` | `172.22.68.254` | `admin` | Router | Web |
| `Mikrotik Sala Generadores` | `192.168.2.1` | `admin` | Router | Web |
| `FreePBX` | `172.22.70.1` | `admin` | VoIP | Web |
| `NVR` (cámaras) | `172.22.80.1-2` | `admin` | Video recorder | Web |
| `Nano Station` | `172.22.90.103` | `ubnt` | Ubiquiti link | Web |
| `PLC Siemens` | `172.22.68.222` | (AD account) | Industrial PLC | Web |
| `PCSISTEMAS01` | `172.22.60.100` | domain user | Sysadmin workstation | RDP |
| `CUPS admin` | `192.168.1.99:631` | `automata` | Printing | Web |
| `Powermeter` | `192.168.2.248` | `admin` | Power meter | Web |

## Flow for reaching a host

1. **VPN up?** If Claude is about to SSH or curl an internal IP, confirm the
   tunnel is active. A blind failure usually means the tunnel is down.
2. **Identify the host** in the inventory above. Note the expected user.
3. **Need credentials?** Ask the user via `AskUserQuestion`:
   > "Password for `sistemas@172.22.100.99` (Histrix)?"
   Use it for the current command. Don't write it to any file, don't echo it
   in a log.
4. **Prefer SSH keys over passwords.** If the user has a key pair set up, their
   `~/.ssh/config` handles auth. Suggest it if they're repeatedly pasting passwords.
5. **Never store credentials in the repo, in CI, or in the `.claude/` folder.**
   Their place is the user's password manager.

## Working with the legacy ERP (Histrix)

This is the most common cross-system task — the RAG pipeline ingests from
Histrix. Typical flows:

- **Inspect data:** SSH to `sistemas@172.22.100.99`, run the ERP's DB client
  locally, export what's needed. Ask for the DB password in the moment.
- **Pull a table for ingestion:** use `pg_dump` / `mysqldump` on the host,
  stream to stdout, pipe through SSH to the SDA workstation or your laptop.
  Never leave a dump file on the legacy server.
- **Schema drift checks:** diff the Histrix schema against the columns
  `services/app/internal/rag/ingest` + the migration-health invariants expect.
  Record any drift as an ADR or a plan entry.

Any time Histrix credentials are about to leave the user's hands (e.g. you're
about to include them in a command), **pause and confirm**. This is the
single highest-blast-radius credential on the network.

## Red flags to catch in reviews

- A `password=`, `secret=`, `token=` literal in any committed file.
- A `.pem` / `.key` / `.ovpn` / `.conf` file with secrets inside the repo tree.
- A script that `echo`s a credential into a log.
- A Dockerfile that `COPY`s a credential file into an image.
- A CI workflow that prints a secret to stdout (use `::add-mask::` or don't print it).

Any of those is a **critical** finding in the `code-review` severity taxonomy.
Fix before merge, rotate the exposed credential.
