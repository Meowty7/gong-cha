# syntax=docker/dockerfile:1.7
# Multi-stage build for the GongchaCUP backend API.
# Build stage: compile a static-ish binary from source.
FROM golang:1.27.1-alpine AS builder
RUN apk add --no-cache git ca-certificates
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath -ldflags="-s -w -X main.version=$(git rev-parse --short HEAD 2>/dev/null || echo dev)" \
    -o /out/server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
    -o /out/healthcheck ./cmd/healthcheck
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
    -o /out/seed ./cmd/seed

# Runtime stage: minimal non-root image with the binary and embedded migrations.
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /
COPY --from=builder /out/server /server
COPY --from=builder /out/healthcheck /healthcheck
COPY --from=builder /out/seed /seed
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/server"]
HEALTHCHECK --interval=10s --timeout=3s --retries=5 --start-period=10s \
    CMD ["/healthcheck"]
