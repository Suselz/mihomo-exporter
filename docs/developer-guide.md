# Developer Guide

## Scope

`mihomo-exporter` collects telemetry from [MetaCubeX/mihomo](https://github.com/MetaCubeX/mihomo) API and exposes Prometheus metrics.

Primary aggregation dimensions:

- client IP (`metadata.sourceIP`)
- outbound (`chains[0]`)

## Runtime flow

1. Poll `/connections`
2. Keep per-connection previous counters by `id`
3. Compute positive deltas
4. Aggregate by client/outbound
5. Poll extra endpoints: `/memory`, `/traffic`, `/version`, `/proxies`, `/rules`
6. Publish metrics via `/metrics`

Notes:

- Counter decreases are treated as reset events.
- Exporter state is in-memory; long history belongs to Prometheus TSDB.

## File map

- Entrypoint: `cmd/mihomo-exporter/main.go`
- Config: `internal/config/config.go`
- Mihomo API client/types:
  - `internal/mihomo/client.go`
  - `internal/mihomo/types.go`
- Delta engine:
  - `internal/engine/delta.go`
  - `internal/engine/aggregate.go`
- Prometheus metrics:
  - `internal/metrics/collector.go`
- HTTP endpoints:
  - `internal/httpserver/server.go`
- Health state:
  - `internal/health/state.go`
- Logging:
  - `internal/logx/logx.go`

## Metrics catalog

### Traffic and connections

- `mihomo_client_upload_bytes_total{client_ip}`
- `mihomo_client_download_bytes_total{client_ip}`
- `mihomo_outbound_upload_bytes_total{outbound}`
- `mihomo_outbound_download_bytes_total{outbound}`
- `mihomo_total_upload_bytes_total`
- `mihomo_total_download_bytes_total`
- `mihomo_client_connections{client_ip}`
- `mihomo_outbound_connections{outbound}`
- `mihomo_connections_total`

### Extra API metrics

From `/memory`:

- `mihomo_memory_inuse_bytes`
- `mihomo_memory_oslimit_bytes`

From `/traffic`:

- `mihomo_traffic_up_bytes_per_second`
- `mihomo_traffic_down_bytes_per_second`
- `mihomo_traffic_upload_total_bytes`
- `mihomo_traffic_download_total_bytes`

From `/version`:

- `mihomo_version_info{version,meta}`

From `/proxies`:

- `mihomo_proxies_total`
- `mihomo_proxies_alive_total`
- `mihomo_proxies_by_type{type}`

From `/rules`:

- `mihomo_rules_total`
- `mihomo_rules_disabled_total`
- `mihomo_rules_hits`
- `mihomo_rules_misses`
- `mihomo_rules_by_type{type}`

### Exporter internals

- `mihomo_exporter_scrape_success`
- `mihomo_exporter_last_scrape_timestamp_seconds`
- `mihomo_exporter_scrape_duration_seconds`
- `mihomo_exporter_parse_errors_total`
- `mihomo_exporter_connection_resets_total`
- `mihomo_exporter_negative_delta_total`
- `mihomo_exporter_tracked_clients`
- `mihomo_exporter_tracked_outbounds`

## Dashboards / PromQL notes

Month usage by client (reset-safe):

- `sum by (client_ip) (increase(mihomo_client_download_bytes_total[30d]))`

Total month usage:

- `sum(increase(mihomo_client_download_bytes_total[30d]))`

Requirements:

- Prometheus TSDB on persistent volume
- retention > 30d

## Tests

Test files:

- `internal/engine/delta_test.go`
- `internal/mihomo/client_test.go`
- `internal/metrics/collector_test.go`

Run:

```bash
go test ./...
```

## CI/CD

Workflow:

- `.github/workflows/release.yml`

Trigger:

- tag push matching `v*`

Actions:

1. Build binaries + checksums
2. Build/push GHCR image (version tags + `latest`)
3. Create GitHub Release and upload artifacts
