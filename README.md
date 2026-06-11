# mihomo-exporter

Prometheus exporter for [MetaCubeX/mihomo](https://github.com/MetaCubeX/mihomo).

The exporter reads Mihomo API endpoints, calculates traffic deltas, and exposes Prometheus metrics for dashboards and alerting.

## What it exports

- Per-client traffic counters
- Per-outbound traffic counters (`outbound = chains[0]`)
- Active connection gauges
- Runtime/API metrics from:
  - `/memory`
  - `/traffic`
  - `/version`
  - `/proxies`
  - `/rules`
- Exporter health metrics

## Quick start

### Local

```bash
go run ./cmd/mihomo-exporter
```

Endpoints:

- `http://127.0.0.1:9109/metrics`
- `http://127.0.0.1:9109/healthz`
- `http://127.0.0.1:9109/readyz`

### Docker

```bash
docker build -t ghcr.io/suselz/mihomo-exporter:latest .
./deploy/docker-run-example.sh
```

### Docker Compose

Current compose example in `deploy/docker-compose/docker-compose.yml` runs exporter container with healthcheck.

```bash
docker compose -f deploy/docker-compose/docker-compose.yml up -d
```

Stop:

```bash
docker compose -f deploy/docker-compose/docker-compose.yml down
```

## Configuration

Environment variables:

- `MIHOMO_URL` (default `http://192.168.0.1:9090`)
- `MIHOMO_CONNECTIONS_PATH` (default `/connections`)
- `EXPORTER_LISTEN_ADDR` (default `:9109`)
- `EXPORTER_METRICS_PATH` (default `/metrics`)
- `EXPORTER_SCRAPE_INTERVAL` (default `2s`)
- `EXPORTER_MAX_CLIENT_SERIES` (default `2000`)
- `EXPORTER_CLIENT_ALLOW_CIDRS` (optional)
- `EXPORTER_LOG_LEVEL` (default `info`, values: `debug|info|warn|error`)

## Main metrics

- `mihomo_client_download_bytes_total{client_ip}`
- `mihomo_client_upload_bytes_total{client_ip}`
- `mihomo_outbound_download_bytes_total{outbound}`
- `mihomo_outbound_upload_bytes_total{outbound}`
- `mihomo_total_download_bytes_total`
- `mihomo_total_upload_bytes_total`
- `mihomo_connections_total`

See full metric catalog and internals in:

- `docs/developer-guide.md`

## Grafana

- Dashboard: `grafana/dashboards/mihomo-exporter-dashboard.json`


