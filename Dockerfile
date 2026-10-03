# syntax=docker/dockerfile:1
FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS build
ARG TARGETOS TARGETARCH VERSION=dev
WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags="-s -w -X github.com/sidimam/wellness-gateway/internal/server.Version=${VERSION}" \
    -o /out/wellness-gateway ./cmd/wellness-gateway

FROM alpine:3.20
LABEL org.opencontainers.image.source="https://github.com/sidimam/wellness-gateway" \
      org.opencontainers.image.description="Always-on booking engine for Technogym mywellness classes (multi-profile) with APNs push for the Wellness Booking iOS app" \
      org.opencontainers.image.licenses="MIT"
RUN apk add --no-cache ca-certificates tzdata curl su-exec
COPY --from=build /out/wellness-gateway /usr/local/bin/wellness-gateway
COPY entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh
ENV LISTEN_ADDR=:8585 CONFIG_DIR=/config TZ=Europe/Rome
EXPOSE 8585
VOLUME ["/config"]
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s \
  CMD curl -fsS http://127.0.0.1:8585/healthz || exit 1
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
