# syntax=docker/dockerfile:1

# ── Build stage ──────────────────────────────────────────────────────────────
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
    -o /bin/container-spy ./cmd/container-spy

# ── Runtime stage ─────────────────────────────────────────────────────────────
FROM alpine:3.21

# openssh-client provides the ssh binary required by Docker's SSH transport.
RUN apk add --no-cache openssh-client ca-certificates

COPY --from=builder /bin/container-spy /bin/container-spy

# Create expected volume mount points.
RUN mkdir -p /config /keys

# Default environment variables.
ENV CONTAINER_SPY_MODE=tui
ENV CONTAINER_SPY_CONFIG=/config/config.yaml

EXPOSE 8080

ENTRYPOINT ["/bin/container-spy"]
