# zero-tunnel

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/aspiand/zero-tunnel)](https://goreportcard.com/report/github.com/aspiand/zero-tunnel)
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/Aspiand/zero-tunnel)

**zero-tunnel** is an automation tool that dynamically manages Cloudflare Tunnel routes based on Docker container labels. It automatically updates tunnel ingress rules and DNS CNAME records when containers start or stop.

> For a detailed technical specification, see [SPECS.md](SPECS.md).

---

## Table of Contents

- [zero-tunnel](#zero-tunnel)
  - [Table of Contents](#table-of-contents)
  - [Features](#features)
  - [How It Works](#how-it-works)
  - [Prerequisites](#prerequisites)
  - [Configuration](#configuration)
    - [Environment Variables](#environment-variables)
    - [Docker Labels](#docker-labels)
    - [Path Matching](#path-matching)
  - [Getting Started](#getting-started)
    - [Docker Compose (Recommended)](#docker-compose-recommended)
    - [Other Installation Methods](#other-installation-methods)
      - [Build from Source](#build-from-source)
      - [Nix](#nix)
  - [Development](#development)
  - [License](#license)

---

## Features

- **Automated Routing** – Expose Docker services via Cloudflare Tunnels using simple container labels.
- **DNS Management** – Automatically creates and removes CNAME records tagged with `managed-by:zero-tunnel`.
- **Real-time Sync** – Updates Cloudflare configuration immediately when containers start or stop.
- **Debounced Updates** – Batches rapid events (e.g., during `docker compose up`) to respect API rate limits.
- **Periodic Reconciliation** – Detects and corrects drift between desired state and actual Cloudflare configuration.

---

## How It Works

`zero-tunnel` continuously bridges your Docker environment and Cloudflare Tunnel configuration:

1. **Startup** – Scans all running containers, builds an initial routing map, and syncs ingress rules and DNS records with Cloudflare.
2. **Event loop** – Watches Docker events (`start`, `stop`, `die`, `destroy`). On `start`, adds or updates the route. On `stop`/`die`, removes it.
3. **Reconciliation** – Runs on a configurable interval to clean up orphaned DNS records and fix any configuration drift caused by external changes.

> [!IMPORTANT]
> `zero-tunnel` manages routing configuration only. `cloudflared` and your target containers must share the same Docker network (or otherwise have direct IP connectivity) so that `cloudflared` can reach them by container name and port.

---

## Prerequisites

- **Docker** – Access to the Docker socket (usually `/var/run/docker.sock`).
- **Cloudflare Account** with:
  - A **Remotely Managed Tunnel** created via the Zero Trust Dashboard.
  - An **API Token** with the following permissions:
    - `Account / Cloudflare Tunnel: Edit`
    - `Account / Account Settings: Read`
    - `Zone / DNS: Edit`
    - `Zone / Zone: Read`

---

## Configuration

### Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `CLOUDFLARE_API_TOKEN` | ✅ | API token with tunnel, DNS, and account permissions. |
| `CLOUDFLARE_ACCOUNT_ID` | ✅ | Your Cloudflare Account ID. |
| `CLOUDFLARE_TUNNEL_ID` | ⚠️ | UUID of an existing tunnel. Mutually exclusive with `CLOUDFLARE_TUNNEL_NAME`. |
| `CLOUDFLARE_TUNNEL_NAME` | ⚠️ | Name of an existing tunnel. Mutually exclusive with `CLOUDFLARE_TUNNEL_ID`. |
| `ZERO_TUNNEL_DEFAULT_DOMAIN` | ❌ | Fallback domain when `zero-tunnel.domain` label is absent. |
| `ZERO_TUNNEL_INTERVAL` | ❌ | Reconciliation interval (default: `300s`). |

> ⚠️ Either `CLOUDFLARE_TUNNEL_ID` or `CLOUDFLARE_TUNNEL_NAME` must be set, but not both.

### Docker Labels

Apply these labels to any container you want to expose.

| Label | Required | Default | Description |
|-------|----------|---------|-------------|
| `zero-tunnel.enable` | ✅ | – | Must be `true` to enable management for this container. |
| `zero-tunnel.subdomain` | ✅ | – | Subdomain prefix for the route (e.g., `myapp`). |
| `zero-tunnel.port` | ✅ | – | Internal container port to forward traffic to. |
| `zero-tunnel.domain` | ❌ | `ZERO_TUNNEL_DEFAULT_DOMAIN` | Root domain (e.g., `example.com`). |
| `zero-tunnel.name` | ❌ | `<hostname>-<container-name>` | Service name used in the tunnel URL. |
| `zero-tunnel.scheme` | ❌ | `http` | Protocol: `http`, `https`, or `tcp`. |
| `zero-tunnel.path` | ❌ | (matches all) | Regex for path-based routing (see [Path Matching](#path-matching)). |
| `zero-tunnel.ephemeral` | ❌ | `true` | When `true`, the DNS record is deleted when the container stops. |

### Path Matching

The `zero-tunnel.path` label uses Cloudflare Tunnel ingress regex syntax:

| Pattern | Matches |
|---------|---------|
| (empty / omitted) | All paths |
| `blog` | Any path containing `blog` (e.g., `/blog`, `/archive/blog/post`) |
| `^/api` | Paths starting with `/api` |
| `\.(jpg\|png\|css\|js)$` | Paths ending with those file extensions |

---

## Getting Started

### Docker Compose (Recommended)

Download the example files:

```bash
wget https://raw.githubusercontent.com/aspiand/zero-tunnel/main/docker-compose.yml
wget -O .env https://raw.githubusercontent.com/aspiand/zero-tunnel/main/.env.example
```

Edit `.env` with your Cloudflare credentials, then start everything:

```bash
docker compose up -d
```

### Other Installation Methods

#### Build from Source

Requires Go to be installed.

```bash
# Build
go build

# Run (ensure environment variables are set)
./zero-tunnel
```

#### Nix

Requires Nix with flakes enabled.

```bash
nix run github:aspiand/zero-tunnel
```

---

## Development

For local development with hot reload, install [`air`](https://github.com/air-verse/air) and run:

```bash
air
```

---

## License

This project is licensed under the **MIT License** – see the [LICENSE](LICENSE) file for details.