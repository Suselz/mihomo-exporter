# mihomo-exporter

Prometheus exporter for Mihomo `/connections` snapshot API.

## Repository structure

```text
.
├── cmd/
│   └── mihomo-exporter/            # application entrypoint
├── internal/                        # exporter implementation
├── grafana/
│   ├── dashboards/                  # dashboard JSON files
│   └── provisioning/                # auto-provisioning for Grafana datasource/dashboards
├── deploy/
│   ├── docker-compose/              # docker-compose example stack
│   ├── docker-run-example.sh        # single-container run example
│   └── prometheus-scrape-example.yml
├── Dockerfile
└── go.mod
```

This keeps exporter code and dashboard assets separated for clean GitHub publishing.

## Features

- Per-client counters:
  - `mihomo_client_upload_bytes_total{client_ip}`
  - `mihomo_client_download_bytes_total{client_ip}`
- Per-outbound counters (`outbound = chains[0]`):
  - `mihomo_outbound_upload_bytes_total{outbound}`
  - `mihomo_outbound_download_bytes_total{outbound}`
- Global counters:
  - `mihomo_total_upload_bytes_total`
  - `mihomo_total_download_bytes_total`
- Active connection gauges by client/outbound and total.
- Exporter health and quality metrics.

## Configuration

Environment variables:

- `MIHOMO_URL` (default `http://192.168.0.1:9090`)
- `MIHOMO_CONNECTIONS_PATH` (default `/connections`)
- `EXPORTER_LISTEN_ADDR` (default `:9109`)
- `EXPORTER_METRICS_PATH` (default `/metrics`)
- `EXPORTER_SCRAPE_INTERVAL` (default `2s`)
- `EXPORTER_MAX_CLIENT_SERIES` (default `2000`)
- `EXPORTER_CLIENT_ALLOW_CIDRS` (optional, comma separated CIDR allowlist)

## Run locally

```bash
go run ./cmd/mihomo-exporter
```

Endpoints:

- `http://127.0.0.1:9109/metrics`
- `http://127.0.0.1:9109/healthz`
- `http://127.0.0.1:9109/readyz`

## Build container

```bash
docker build -t ghcr.io/suselz/mihomo-exporter:latest .
```

Single-container run example:

```bash
./deploy/docker-run-example.sh
```

## Docker Compose example (Exporter + Prometheus + Grafana)

Example stack files:

- `deploy/docker-compose/docker-compose.yml`
- `deploy/docker-compose/prometheus.yml`

Run stack:

```bash
docker compose -f deploy/docker-compose/docker-compose.yml up -d --build
```

Access:

- Exporter metrics: `http://127.0.0.1:9109/metrics`
- Prometheus UI: `http://127.0.0.1:9091`
- Grafana UI: `http://127.0.0.1:3000` (default `admin/admin`)

Grafana auto-loads:

- datasource from `grafana/provisioning/datasources/prometheus.yml`
- dashboard from `grafana/dashboards/mihomo-exporter-dashboard.json`

Stop stack:

```bash
docker compose -f deploy/docker-compose/docker-compose.yml down
```

## Prometheus and Grafana files

- Prometheus scrape sample: `deploy/prometheus-scrape-example.yml`
- Grafana dashboard JSON: `grafana/dashboards/mihomo-exporter-dashboard.json`

## Release automation (GitHub tags)

This repository includes release workflow:

- `.github/workflows/release.yml`

On tag push matching `v*` (for example `v1.0.0`) workflow does:

1. Build multi-platform binaries (`linux`, `darwin`, `windows`).
2. Generate `dist/checksums.txt`.
3. Build and push multi-arch container image to GHCR:
   - `ghcr.io/<owner>/<repo>:<tag>`
4. Create GitHub Release and attach binaries + checksums.

Create release tag:

```bash
git tag v1.0.0
git push origin v1.0.0
```

## Notes on data model

- Exporter polls `/connections` on interval and computes positive deltas.
- Negative deltas are treated as resets and not added to counters.
- For outbound attribution, only `chains[0]` is used by design.
