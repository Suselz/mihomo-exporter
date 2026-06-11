#!/usr/bin/env sh
set -eu

docker run -d \
  --name mihomo-exporter \
  --restart unless-stopped \
  --read-only \
  --tmpfs /tmp:rw,nosuid,nodev,size=16m \
  --cpus 0.50 \
  --memory 128m \
  -p 9109:9109 \
  -e MIHOMO_URL=http://192.168.0.1::9090 \
  -e MIHOMO_CONNECTIONS_PATH=/connections \
  -e EXPORTER_LISTEN_ADDR=:9109 \
  -e EXPORTER_METRICS_PATH=/metrics \
  -e EXPORTER_SCRAPE_INTERVAL=2s \
  -e EXPORTER_MAX_CLIENT_SERIES=2000 \
  -e EXPORTER_LOG_LEVEL=info \
  ghcr.io/suselz/mihomo-exporter:latest
