# ── Build stage ─────────────────────────────────────────────────────────────
FROM golang:1.24 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o bajet_server .

# ── Runtime stage ────────────────────────────────────────────────────────────
FROM debian:bookworm-slim
RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=builder /app/bajet_server .

VOLUME ["/data"]
ENV DB_PATH=/data/bajet.db
ENV PORT=8080
EXPOSE 8080

CMD ["./bajet_server"]
