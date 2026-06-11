# syntax=docker/dockerfile:1.7

FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags='-s -w' -o /out/mihomo-exporter ./cmd/mihomo-exporter

FROM gcr.io/distroless/static-debian12:nonroot
ENV EXPORTER_LISTEN_ADDR=:9109 \
    EXPORTER_METRICS_PATH=/metrics \
    EXPORTER_SCRAPE_INTERVAL=2s \
    MIHOMO_URL=http://192.168.0.1:9090 \
    MIHOMO_CONNECTIONS_PATH=/connections \
    EXPORTER_MAX_CLIENT_SERIES=2000
COPY --from=builder /out/mihomo-exporter /mihomo-exporter
EXPOSE 9109
ENTRYPOINT ["/mihomo-exporter"]
