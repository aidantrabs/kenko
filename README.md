# Kenko

A health monitoring service that periodically checks HTTP endpoints and reports their status. Built with Go, load-balanced with NGINX, backed by Redis for shared state, and observable via Prometheus and Grafana.

## Quickstart

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
