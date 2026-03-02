# kenko

[![ci](https://github.com/aidantrabs/kenko/actions/workflows/ci.yml/badge.svg)](https://github.com/aidantrabs/kenko/actions/workflows/ci.yml)
[![go reference](https://pkg.go.dev/badge/github.com/aidantrabs/kenko.svg)](https://pkg.go.dev/github.com/aidantrabs/kenko)
[![go report card](https://goreportcard.com/badge/github.com/aidantrabs/kenko)](https://goreportcard.com/report/github.com/aidantrabs/kenko)
[![license: mit](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

a health monitoring sdk and standalone service for go. drop health-check monitoring into your own app with a few lines of code, or run the full docker compose stack with nginx, redis, prometheus, and grafana.

## install

```bash
go get github.com/aidantrabs/kenko@latest
```

optional sub-packages:

```bash
go get github.com/aidantrabs/kenko/redisstore   # redis-backed state
go get github.com/aidantrabs/kenko/prommetrics   # prometheus metrics
```

## usage

### minimal example

```go
package main

import (
    "context"
    "net/http"
    "time"

    "github.com/aidantrabs/kenko"
)

func main() {
    k, err := kenko.New(
        kenko.WithTarget("api", "https://api.example.com/health"),
        kenko.WithTarget("db", "https://db.example.com/health"),
        kenko.WithInterval(30 * time.Second),
    )
    if err != nil {
        panic(err)
    }

    mux := http.NewServeMux()
    k.RegisterHandlers(mux)

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    go k.Run(ctx)

    http.ListenAndServe(":8080", mux)
}
```

this gives you `/health`, `/ready`, and `/status` endpoints using an in-memory store with zero external dependencies.

### with redis and prometheus

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
k.RegisterHandlers(mux)
go k.Run(ctx)
```

### low-level api

use the low-level api if you want direct access to check results without http handlers:

```go
checker, _ := kenko.NewChecker(
    kenko.WithTarget("api", "https://api.example.com"),
    kenko.WithInterval(30 * time.Second),
)
go checker.Run(ctx)
results, _ := checker.Results()
```

the root package has zero third-party dependencies. redis and prometheus are opt-in via sub-packages.

## standalone quickstart

```bash
docker compose up --build
```

this starts 3 monitor instances behind nginx, plus redis, prometheus, and grafana.

## endpoints

| endpoint   | description                                      | example                  |
|------------|--------------------------------------------------|--------------------------|
| `/health`  | liveness probe — checks service and dependencies | `curl localhost/health`  |
| `/ready`   | readiness probe — 503 until first check cycle    | `curl localhost/ready`   |
| `/status`  | detailed status of all monitored targets         | `curl localhost/status`  |
| `/metrics` | prometheus metrics                               | `curl localhost/metrics` |

## configuration

edit `configs/config.yaml`:

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

| field            | description                          | default       |
|------------------|--------------------------------------|---------------|
| `port`           | http server port (1-65535)           | `6969`        |
| `check_interval` | time between check cycles            | `30s`         |
| `check_timeout`  | timeout per http check               | `5s`          |
| `redis_addr`     | redis address (host:port)            | `redis:6379`  |
| `targets`        | list of endpoints to monitor         | —             |
| `targets[].name` | display name for the target          | —             |
| `targets[].url`  | url to check (must be valid http(s)) | —             |

## architecture

```
                         ┌─────────────────────────────────────┐
                         │           github actions            │
                         │  (build, test, push docker images)  │
                         └──────────────────┬──────────────────┘
                                            │
                                            ▼
┌──────────┐    ┌─────────────────────────────────────────────────────────┐
│  users   │───▶│                    nginx (load balancer)               │
└──────────┘    └─────────────────────────┬───────────────────────────────┘
                                          │
                    ┌─────────────────────┼─────────────────────┐
                    ▼                     ▼                     ▼
            ┌──────────────┐      ┌──────────────┐      ┌──────────────┐
            │  monitor     │      │  monitor     │      │  monitor     │
            │  instance 1  │      │  instance 2  │      │  instance 3  │
            │  (go)        │      │  (go)        │      │  (go)        │
            └──────┬───────┘      └──────┬───────┘      └──────┬───────┘
                   │                     │                     │
                   └─────────────────────┼─────────────────────┘
                                         │
                    ┌────────────────────┼────────────────────┐
                    ▼                    ▼                    ▼
            ┌──────────────┐     ┌──────────────┐     ┌──────────────┐
            │    redis     │     │  prometheus  │     │   grafana    │
            │   (state)    │     │  (metrics)   │     │ (dashboards) │
            └──────────────┘     └──────────────┘     └──────────────┘
```

## services

| service    | port  | description               |
|------------|-------|---------------------------|
| nginx      | 80    | load balancer / proxy     |
| prometheus | 9090  | metrics collection        |
| grafana    | 3002  | dashboards (anonymous)    |
| kenko      | 6969  | monitor (internal only)   |
| redis      | 6379  | shared state (internal)   |

## docker compose verification

after starting the stack with `docker compose up --build -d`, verify all endpoints:

```bash
# liveness — should return {"status":"healthy","redis":"up"}
curl -s localhost/health | jq .

# readiness — 503 initially, 200 after first check cycle
curl -s -o /dev/null -w '%{http_code}' localhost/ready

# target status — detailed per-target results
curl -s localhost/status | jq .

# prometheus — check kenko metrics are being scraped
curl -s 'localhost:9090/api/v1/query?query=kenko_target_up' | jq .

# grafana — health check
curl -s -o /dev/null -w '%{http_code}' localhost:3002/api/health
```

## development

```bash
make build       # build the binary
make test        # run tests
make test-cover  # run tests with coverage
make lint        # run linter
make run         # run locally
make docker-up   # start all services
make docker-down # stop all services
make clean       # remove build artifacts
```

## license

[MIT](LICENSE)
