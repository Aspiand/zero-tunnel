# zero-tunnel

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/aspiand/zero-tunnel)](https://goreportcard.com/report/github.com/aspiand/zero-tunnel)
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/Aspiand/zero-tunnel)

**zero-tunnel** is an automation tool that dynamically manages Cloudflare Tunnel routes based on Docker container labels. It automatically updates tunnel ingress rules and DNS CNAME records when containers start or stop.

---

## Table of Contents

- [zero-tunnel](#zero-tunnel)
  - [Table of Contents](#table-of-contents)
  - [Features](#features)
  - [Prerequisites](#prerequisites)
  - [Configuration](#configuration)
    - [Environment Variables](#environment-variables)
    - [Docker Labels](#docker-labels)
      - [Path Matching](#path-matching)
  - [Getting Started](#getting-started)
    - [⚠️ Network Requirement](#️-network-requirement)
    - [Full Example with Docker Compose](#full-example-with-docker-compose)
  - [Development \& Testing](#development--testing)
    - [Running Tests](#running-tests)
  - [License](#license)

---

## Features

- **Automated Routing** – Expose Docker services via Cloudflare Tunnels using simple labels.
- **DNS Management** – Automatically creates and updates CNAME records with `managed-by:zero-tunnel` comments.
- **Real-time Sync** – Updates Cloudflare configuration immediately when containers start or stop.
- **Debounced Updates** – Batches multiple events (e.g., during `docker-compose up`) to respect API rate limits.
- **Periodic Reconciliation** – Ensures the desired state is maintained even after external changes.

---

## Prerequisites

- **Docker** – Access to the Docker socket (usually `/var/run/docker.sock`).
- **Network** – `zero-tunnel` and target containers must be in the **same Docker Network** (or reachable) so `cloudflared` can route traffic.
- **Cloudflare Account** with:
  - A **Remotely Managed Tunnel** created via the Zero Trust Dashboard.
  - An **API Token** having these permissions:
    - Account / Cloudflare Tunnel: `Edit`
    - Account / Account Settings: `Read`
    - Zone / DNS: `Edit`
    - Zone / Zone: `Read`

---

## Configuration

### Environment Variables

| Variable | Description |
|----------|-------------|
| `CLOUDFLARE_API_TOKEN` | Your Cloudflare API Token (required). |
| `CLOUDFLARE_ACCOUNT_ID` | Your Cloudflare Account ID (required). |
| `CLOUDFLARE_TUNNEL_ID` | UUID of the existing tunnel (mutually exclusive with `NAME`). |
| `CLOUDFLARE_TUNNEL_NAME` | Name of the existing tunnel (mutually exclusive with `ID`). |
| `ZERO_TUNNEL_DEFAULT_DOMAIN` | Default domain used when a container label lacks `domain`. |
| `ZERO_TUNNEL_INTERVAL` | Reconciliation interval, e.g., `300s` (default: `300s`). |

### Docker Labels

Apply these labels to containers you want to expose.

| Label | Required | Example | Description |
|-------|----------|---------|-------------|
| `zero-tunnel.enable` | ✅ | `true` | Must be `true` to enable management. |
| `zero-tunnel.subdomain` | ✅ | `myapp` | Subdomain for the route. |
| `zero-tunnel.port` | ✅ | `8080` | Internal container port. |
| `zero-tunnel.domain` | ❌ | `example.com` | Root domain (falls back to `ZERO_TUNNEL_DEFAULT_DOMAIN`). |
| `zero-tunnel.name` | ❌ | `web-server` | Target service name (default: `<host-hostname>-<container-name>`). |
| `zero-tunnel.scheme` | ❌ | `http` | `http`, `https`, or `tcp` (default: `http`). |
| `zero-tunnel.path` | ❌ | `^/api` | Path matching regex (see [Path Matching](#path-matching)). |
| `zero-tunnel.ephemeral` | ❌ | `false` | If `true` (default), DNS is deleted when container stops. |

#### Path Matching

The `zero-tunnel.path` field supports regex patterns as defined by Cloudflare Tunnel ingress rules:

| Pattern | Matches |
|---------|---------|
| (empty / omitted) | all paths |
| `blog` | `/blog`, `/archive/blog/post`, etc. |
| `^/api` | anything starting with `/api` |
| `\.(jpg\|png\|css\|js)$` | files with those extensions |

---

## Getting Started

### ⚠️ Network Requirement

Your target application (e.g., `nginx-test`) **and** the `cloudflared` container **must be attached to the same Docker network**.  
`zero-tunnel` only manages routes – connectivity between `cloudflared` and your app is your responsibility.

### Full Example with Docker Compose

Create a `docker-compose.yml` file:

```yaml
services:
  zero-tunnel:
    container_name: zero-tunnel
    image: ghcr.io/aspiand/zero-tunnel:latest
    restart: unless-stopped
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
    env_file: .env

  cloudflared:
    image: cloudflare/cloudflared:latest
    container_name: cloudflared
    restart: unless-stopped
    command: tunnel run
    environment:
      - TUNNEL_TOKEN=${CLOUDFLARE_TUNNEL_TOKEN}

  nginx-test:
    image: nginx:alpine
    container_name: nginx-test
    labels:
      zero-tunnel.enable: "true"
      zero-tunnel.subdomain: "nginx"
      zero-tunnel.port: "80"
      # zero-tunnel.domain: "example.com"   # optional, falls back to ZERO_TUNNEL_DEFAULT_DOMAIN

networks:
  default:
    name: zero-tunnel
```

Then start everything:

```bash
docker-compose up -d
```

> 💡 **How it works**  
> `cloudflared` and `nginx-test` share same network.
> `zero-tunnel` detects the `nginx-test` labels and tells Cloudflare to route `nginx.yourdomain.com` → `http://nginx-test:80` (container name and port).  
> `cloudflared` receives this configuration and proxies traffic directly to `nginx-test` inside the same network.

---

## Development & Testing

### Running Tests

```bash
go test -v ./...
```

**E2E tests** require a running Docker daemon. They automatically:
1. Start a mock Cloudflare API server.
2. Pull a small `busybox` image.
3. Create and manage temporary containers.
4. Verify the full lifecycle of tunnel and DNS management.

---

## License

This project is licensed under the **MIT License** – see the [LICENSE](LICENSE) file for details.