# Kenko

A health monitoring SDK and standalone service for Go. Drop health-check monitoring into your own app with a few lines of code, or run the full Docker Compose stack with NGINX, Redis, Prometheus, and Grafana.

## SDK Usage

### Quick setup (high-level API)

```go
import (
    "github.com/aidantrabs/kenko"
    "github.com/aidantrabs/kenko/redisstore"
    "github.com/aidantrabs/kenko/prommetrics"
)

k, _ := kenko.New(
    kenko.WithTarget("api", "https://api.example.com"),
    kenko.WithTarget("db", "https://db.example.com/health"),
    kenko.WithInterval(30 * time.Second),
    kenko.WithStore(redisstore.New("localhost:6379")),
    kenko.WithMetrics(prommetrics.New()),
)

mux := http.NewServeMux()
k.RegisterHandlers(mux) // adds /health, /ready, /status
go k.Run(ctx)
```

### Low-level API

```go
checker, _ := kenko.NewChecker(
    kenko.WithTarget("api", "https://api.example.com"),
    kenko.WithInterval(30 * time.Second),
)
go checker.Run(ctx)
results, _ := checker.Results()
```

The root package has zero third-party dependencies. Redis and Prometheus are opt-in via sub-packages.

## Standalone Quickstart

```bash
docker compose up --build
```

This starts 3 monitor instances behind NGINX, plus Redis, Prometheus, and Grafana.

## Endpoints

| Endpoint   | Description                                      | Example                  |
|------------|--------------------------------------------------|--------------------------|
| `/health`  | Liveness probe — checks service and dependencies | `curl localhost/health`  |
| `/ready`   | Readiness probe — 503 until first check cycle    | `curl localhost/ready`   |
| `/status`  | Detailed status of all monitored targets         | `curl localhost/status`  |
| `/metrics` | Prometheus metrics                               | `curl localhost/metrics` |

## Configuration

Edit `configs/config.yaml`:

```yaml
port: 6969
check_interval: 30s
check_timeout: 5s
redis_addr: redis:6379

targets:
  - name: google
    url: https://www.google.com
  - name: github
    url: https://github.com
```

| Field            | Description                          | Default       |
|------------------|--------------------------------------|---------------|
| `port`           | HTTP server port (1-65535)           | `6969`        |
| `check_interval` | Time between check cycles            | `30s`         |
| `check_timeout`  | Timeout per HTTP check               | `5s`          |
| `redis_addr`     | Redis address (host:port)            | `redis:6379`  |
| `targets`        | List of endpoints to monitor         | —             |
| `targets[].name` | Display name for the target          | —             |
| `targets[].url`  | URL to check (must be valid HTTP(S)) | —             |

## Architecture

```
                         ┌─────────────────────────────────────┐
                         │           GitHub Actions            │
                         │  (build, test, push Docker images)  │
                         └──────────────────┬──────────────────┘
                                            │
                                            ▼
┌──────────┐    ┌─────────────────────────────────────────────────────────┐
│  Users   │───▶│                    NGINX (Load Balancer)                │
└──────────┘    └─────────────────────────┬───────────────────────────────┘
                                          │
                    ┌─────────────────────┼─────────────────────┐
                    ▼                     ▼                     ▼
            ┌──────────────┐      ┌──────────────┐      ┌──────────────┐
            │  Monitor     │      │  Monitor     │      │  Monitor     │
            │  Instance 1  │      │  Instance 2  │      │  Instance 3  │
            │  (Go)        │      │  (Go)        │      │  (Go)        │
            └──────┬───────┘      └──────┬───────┘      └──────┬───────┘
                   │                     │                     │
                   └─────────────────────┼─────────────────────┘
                                         │
                    ┌────────────────────┼────────────────────┐
                    ▼                    ▼                    ▼
            ┌──────────────┐     ┌──────────────┐     ┌──────────────┐
            │    Redis     │     │  Prometheus  │     │   Grafana    │
            │   (state)    │     │  (metrics)   │     │ (dashboards) │
            └──────────────┘     └──────────────┘     └──────────────┘
```

## Services

| Service    | Port  | Description               |
|------------|-------|---------------------------|
| NGINX      | 80    | Load balancer / proxy     |
| Prometheus | 9090  | Metrics collection        |
| Grafana    | 3002  | Dashboards (anonymous)    |
| Kenko      | 6969  | Monitor (internal only)   |
| Redis      | 6379  | Shared state (internal)   |

## Development

```bash
make build       # Build the binary
make test        # Run tests
make lint        # Run linter
make run         # Run locally
make docker-up   # Start all services
make docker-down # Stop all services
make clean       # Remove build artifacts
```

## License

[MIT](LICENSE)
