# Specification: zero-tunnel

**zero-tunnel** automates Cloudflare Tunnel route management based on Docker container labels. It monitors Docker events and continuously syncs tunnel ingress rules and DNS CNAME records.

> **Important**: This tool assumes the Cloudflare Tunnel connector (`cloudflared`) and the target containers share a network or have direct IP connectivity. `zero-tunnel` manages routing logic only – not Docker network topology.

---

## Table of Contents

- [Specification: zero-tunnel](#specification-zero-tunnel)
  - [Table of Contents](#table-of-contents)
  - [1. Objectives](#1-objectives)
  - [2. Core Components](#2-core-components)
    - [2.1 Watcher (Docker)](#21-watcher-docker)
    - [2.2 Provider (Cloudflare)](#22-provider-cloudflare)
    - [2.3 Engine](#23-engine)
  - [3. Configuration \& Labels](#3-configuration--labels)
    - [3.1 Global Configuration](#31-global-configuration)
    - [3.2 Docker Labels](#32-docker-labels)
    - [3.3 Path Matching](#33-path-matching)
    - [3.4 Ephemeral DNS Cleanup](#34-ephemeral-dns-cleanup)
  - [4. Technical Implementation](#4-technical-implementation)
    - [4.1 Concurrency Model](#41-concurrency-model)
    - [4.2 Workflow](#42-workflow)
    - [4.3 Cloudflare Ingress Rules](#43-cloudflare-ingress-rules)
  - [5. Future Work](#5-future-work)

---

## 1. Objectives

- **Automated Routing** – Expose Docker services via Cloudflare Tunnels using only labels.
- **Dynamic Configuration** – Update tunnel ingress rules in real-time when containers start or stop.
- **DNS Management** – Automatically create and remove CNAME records pointing to the Cloudflare Tunnel.
- **Ease of Use** – Minimal configuration; most settings derived from container metadata.

---

## 2. Core Components

### 2.1 Watcher (Docker)
- Monitors Docker events: `start`, `stop`, `die`, `destroy`.
- Periodically polls all running containers (on startup and for reconciliation).
- Extracts configuration from container labels.

### 2.2 Provider (Cloudflare)
- Interacts with Cloudflare API v4 (via `cloudflare-go/v6`).
- Updates **Remotely Managed Tunnel** configuration (Ingress Rules).
- Manages DNS CNAME records for configured hostnames.
- *MVP scope*: manage routes for an existing tunnel only.

### 2.3 Engine
- Orchestrates Watcher and Provider.
- Maintains internal state to ensure tunnel configuration consistency.
- Handles concurrency and API rate limiting.

---

## 3. Configuration & Labels

### 3.1 Global Configuration

All options can be set via environment variables or a config file (using Viper).

| Environment Variable | Required | Description |
|----------------------|----------|-------------|
| `CLOUDFLARE_API_TOKEN` | ✅ | API token with permissions: Tunnel:Edit, Account:Read, DNS:Edit, Zone:Read. |
| `CLOUDFLARE_ACCOUNT_ID` | ✅ | Cloudflare Account ID. |
| `CLOUDFLARE_TUNNEL_ID` | ⚠️ | UUID of existing tunnel (mutually exclusive with `NAME`). |
| `CLOUDFLARE_TUNNEL_NAME` | ⚠️ | Name of existing tunnel (mutually exclusive with `ID`). |
| `ZERO_TUNNEL_DEFAULT_DOMAIN` | ❌ | Fallback domain if label missing. |
| `ZERO_TUNNEL_INTERVAL` | ❌ | Reconciliation interval (default: `300s`). |

### 3.2 Docker Labels

| Label | Required | Default | Description |
|-------|----------|---------|-------------|
| `zero-tunnel.enable` | ✅ | – | Must be `true` to process this container. |
| `zero-tunnel.subdomain` | ✅ | – | Subdomain for the route. |
| `zero-tunnel.domain` | ❌ | `ZERO_TUNNEL_DEFAULT_DOMAIN` | Root domain. |
| `zero-tunnel.port` | ✅ | – | Internal container port. |
| `zero-tunnel.name` | ❌ | `<host-hostname>-<container-name>` | Target service name used in the tunnel URL. |
| `zero-tunnel.scheme` | ❌ | `http` | Protocol: `http`, `https`, or `tcp`. |
| `zero-tunnel.path` | ❌ | (empty → matches all) | Regex for path matching. |
| `zero-tunnel.ephemeral` | ❌ | `true` | If `true`, DNS record is deleted when container stops. |

### 3.3 Path Matching

Uses Cloudflare Tunnel ingress regex syntax:

- **Empty / omitted** → matches all paths.
- `blog` → matches any path containing `blog` (e.g., `/blog`, `/archive/blog/post`).
- `^/api` → matches paths starting with `/api`.
- `\.(jpg\|png\|css\|js)$` → matches file extensions.

### 3.4 Ephemeral DNS Cleanup

- **Default**: All managed routes are ephemeral. When a container stops, its associated DNS record is deleted.
- **Identification**: Every DNS record created/updated by zero-tunnel includes the comment `managed-by:zero-tunnel`.
- **Override**: Set `zero-tunnel.ephemeral=false` to keep the DNS record even when the container is offline.
- **Cleanup**: During startup and each reconciliation cycle, zero-tunnel deletes orphaned DNS records (ones with the `managed-by:zero-tunnel` comment that have no corresponding active ephemeral container).

---

## 4. Technical Implementation

### 4.1 Concurrency Model

- **Serialized Updates**: A single worker/queue processes Cloudflare updates because the tunnel configuration is a single blob that must be replaced atomically.
- **Cancellation**: Uses `context.Context` for graceful shutdown and request cancellation.
- **Logging**: Structured logging with `slog`.

### 4.2 Workflow

1. **Startup**
   - List all running containers.
   - Build initial routing map.
   - Sync with Cloudflare: update ingress rules + ensure DNS records.

2. **Event Loop**
   - On `start` → add/update route for the container.
   - On `die`/`stop` → mark route for removal.
   - **Debounce**: Batch events (e.g., during `docker-compose up`) to avoid excessive API calls.

3. **Reconciliation** (periodic)
   - Check if current Cloudflare configuration matches desired state.
   - Clean up orphaned ephemeral DNS records.

### 4.3 Cloudflare Ingress Rules

The tool maintains an ingress list with the following structure:

```
1. Container-specific rules (e.g., myapp.example.com → http://container-name:8080)
2. ... (more container rules)
3. Catch-all rule: http_status:404
```

The catch-all rule is mandatory and always placed at the end.

---

## 5. Future Work

- **Automated Tunnel Provisioning** – Create a new tunnel if neither ID nor Name is provided.
- **Local Configuration Provider** – Support writing to `config.yaml` for locally managed tunnels.
- **Multiple Tunnels** – Distribute routes across multiple Cloudflare Tunnels based on labels.

---
