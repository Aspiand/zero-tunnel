# zero-tunnel

zero-tunnel is an automation tool designed to manage Cloudflare Tunnel routes dynamically based on Docker container labels. It automatically updates tunnel ingress rules and manages DNS CNAME records as containers start and stop.

## Features

- **Automated Routing**: Expose Docker services via Cloudflare Tunnels using simple labels.
- **DNS Management**: Automatically creates and updates CNAME records with `managed-by:zero-tunnel` comments.
- **Real-time Sync**: Updates Cloudflare configuration immediately when containers start or stop.
- **Debounced Updates**: Batch-processes multiple events (e.g., during `docker-compose up`) to stay within API rate limits.
- **Periodic Reconciliation**: Ensures the desired state is maintained even if external changes occur.

## Prerequisites

- **Docker**: Access to the Docker Socket (usually `/var/run/docker.sock`).
- **Cloudflare Account**:
    - A Remotely Managed Tunnel created via the Zero Trust Dashboard.
    - An API Token with the following permissions:
        - Account / Cloudflare Tunnel: Edit
        - Account / Account Settings: Read
        - Zone / DNS: Edit
        - Zone / Zone: Read

## Configuration

| Environment Variable | Description |
|----------------------|-------------|
| CLOUDFLARE_API_TOKEN | Your Cloudflare API Token. |
| CLOUDFLARE_ACCOUNT_ID | Your Cloudflare Account ID. |
| CLOUDFLARE_TUNNEL_ID | The UUID of your existing Tunnel. |
| ZERO_TUNNEL_DEFAULT_DOMAIN | (Optional) Default domain to use if `zero-tunnel.domain` label is missing. |
| ZERO_TUNNEL_INTERVAL | (Optional) Reconciliation interval (e.g., `300s`). |

## Getting Started

### Using Docker

The easiest way to run zero-tunnel is using the official image from GitHub Container Registry:

```yaml
# docker-compose.yml
services:
  zero-tunnel:
    image: ghcr.io/aspiand/zero-tunnel:latest
    container_name: zero-tunnel
    restart: always
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
    environment:
      - CLOUDFLARE_API_TOKEN=your_token
      - CLOUDFLARE_ACCOUNT_ID=your_account_id
      - CLOUDFLARE_TUNNEL_ID=your_tunnel_id
      - ZERO_TUNNEL_DEFAULT_DOMAIN=example.com
```

## Docker Labels

To expose a container, add the following labels:

| Label | Description | Example |
|-------|-------------|---------|
| `zero-tunnel.enable` | Set to `true` to enable management. | `true` |
| `zero-tunnel.subdomain` | The subdomain to expose. | `myapp` |
| `zero-tunnel.domain` | The root domain (falls back to `ZERO_TUNNEL_DEFAULT_DOMAIN`). | `example.com` |
| `zero-tunnel.port` | The internal container port. | `8080` |
| `zero-tunnel.name` | (Optional) Target service name. | `web-server` |
| `zero-tunnel.scheme` | (Optional) `http`, `https`, or `tcp`. Default: `http`. | `http` |
| `zero-tunnel.path` | (Optional) Path matching regex. | `^/api` |
| `zero-tunnel.ephemeral` | (Optional) If `true`, DNS is deleted when container stops. Default: `true`. | `false` |

## License
This project is licensed under the MIT License - see the LICENSE file for details.
