# Specification: zero-tunnel

`zero-tunnel` is an automation tool designed to manage Cloudflare Tunnel routes dynamically based on Docker container labels. It functions similarly to Watchtower, monitoring the Docker daemon for container lifecycle events and automatically updating Cloudflare Tunnel ingress rules and DNS records.

## 1. Objectives

- **Automated Routing**: Automatically expose Docker services via Cloudflare Tunnels using labels.
- **Dynamic Configuration**: Update tunnel ingress rules in real-time when containers start or stop.
- **DNS Management**: Automatically create and remove CNAME records pointing to the Cloudflare Tunnel.
- **Ease of Use**: Minimal configuration required; most settings are derived from container metadata.

## 2. Core Components

### 2.1 Watcher (Docker)
- Monitors Docker events (`start`, `stop`, `die`, `destroy`).
- Periodically polls all running containers (on startup or for reconciliation).
- Extracts configuration from labels.

### 2.2 Provider (Cloudflare)
- Interacts with Cloudflare API v4 (via `cloudflare-go/v6`).
- Updates **Remotely Managed Tunnels** configuration (Ingress Rules).
- Manages DNS CNAME records for the configured hostnames.
- *Note: MVP focuses on managing routes for an existing tunnel.*

### 2.3 Engine
- Orchestrates the flow between Watcher and Provider.
- Manages state to ensure the tunnel configuration is consistent.
- Handles concurrency and rate-limiting for API calls.

## 3. Configuration & Labels

### 3.1 Global Configuration (Viper)
Configurable via environment variables or a config file:
- `CLOUDFLARE_API_TOKEN`: API Token with the following permissions:
    - **Account / Cloudflare Tunnel**: `Edit`
    - **Account / Account Settings**: `Read`
    - **Zone / DNS**: `Edit`
    - **Zone / Zone**: `Read`
- `CLOUDFLARE_ACCOUNT_ID`: Cloudflare Account ID.
- `CLOUDFLARE_TUNNEL_ID`: The UUID of the existing remotely managed tunnel.
- `ZERO_TUNNEL_DEFAULT_DOMAIN`: (Optional) Default domain if not specified in labels.
- `ZERO_TUNNEL_INTERVAL`: (Optional) Reconciliation interval (default: 300s).

### 3.2 Docker Labels
| Label | Description | Example |
|-------|-------------|---------|
| `zero-tunnel.enable` | Must be `true` for the container to be processed. | `true` |
| `zero-tunnel.name` | (Optional) Overrides the target service name in the Cloudflare service URL. Default: `<host-hostname>-<container-name>`. | `my-server-gallery` |
| `zero-tunnel.subdomain` | The subdomain to use. | `gallery` |
| `zero-tunnel.domain` | The root domain. | `aspian.my.id` |
| `zero-tunnel.port` | Internal port of the container. | `8080` |
| `zero-tunnel.scheme` | Protocol to use (`http`, `https`, `tcp`). | `http` |
| `zero-tunnel.path` | (Optional) Path for the ingress rule. | `/api` |
| `zero-tunnel.ephemeral` | (Optional) If `true`, DNS is deleted when container stops. Default: `true`. | `true` |

### 3.3 Path Matching
The `zero-tunnel.path` field supports regex patterns as defined by Cloudflare Tunnel ingress rules:
- **Match all paths**: Leave the label empty or omit it.
- **Match anywhere in path**: `blog` (matches `/blog`, `/archive/blog/post`, etc.)
- **Match path prefix**: `^/api`
- **Match files by extension**: `\.(jpg|png|css|js)$`

### 3.4 Ephemeral DNS (Cleanup)
`zero-tunnel` manages the lifecycle of DNS records it creates, following an ephemeral-by-default pattern:
- **Default Behavior**: All managed routes are ephemeral. When a container stops or is removed, its associated DNS record is deleted.
- **Identification**: Every DNS record created or updated by the tool includes the comment `managed-by:zero-tunnel`.
- **Label Control**: Users can set `zero-tunnel.ephemeral=false` to keep the DNS record even if the container is offline.
- **Automatic Removal**: During each synchronization cycle (Startup and Reconciliation), the tool identifies DNS records containing the `managed-by:zero-tunnel` comment that no longer have a corresponding active container (and are marked as ephemeral). These orphaned records are automatically deleted from Cloudflare.

## 4. Technical Implementation

### 4.1 Concurrency Model
- Use a single worker or a serialized queue to update Cloudflare configuration to avoid race conditions (Cloudflare Tunnel config is a single blob that must be replaced entirely).
- Use `context.Context` for graceful shutdown and request cancellation.
- Use `slog` for structured logging.

### 4.2 Workflow
1. **Startup**:
   - List all containers and build an initial routing map.
   - Sync the routing map with Cloudflare (Update Ingress + Ensure DNS).
2. **Event Loop**:
   - On `start`: Add/Update route for the container.
   - On `die`/`stop`: Mark route for removal.
   - Debounce updates to avoid multiple API calls during a `docker-compose up` sequence.
3. **Reconciliation**:
   - Periodically check if the current Cloudflare configuration matches the desired state.

### 4.3 Cloudflare Ingress Rules
The tool will maintain the ingress list:
1. Container-specific rules (e.g., `idk.aspian.my.id -> http://<host-hostname>-<container-name>:8080`).
2. A mandatory catch-all rule at the end (`http_status:404`).

## 5. Future Work
- **Lookup Tunnel ID by Name**: Allow users to provide a tunnel name instead of a UUID, with the tool performing an automatic lookup at startup.
- **Automated Tunnel Provisioning**: Automatically create a new tunnel if `CLOUDFLARE_TUNNEL_ID` is not provided, including credential management.
- **Local Configuration Provider**: Support for writing to `config.yaml` for locally managed tunnels.
- **Multiple Tunnels**: Ability to distribute routes across multiple Cloudflare Tunnels based on labels.

## 6. References
- [Watchtower](https://github.com/containrrr/watchtower) for Docker event monitoring patterns.
- [Cloudflare Go SDK v6](https://github.com/cloudflare/cloudflare-go) for API interactions.
