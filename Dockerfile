# --- Build stage ---
FROM golang:1.24-bookworm AS builder

WORKDIR /build

COPY go.mod go.sum ./
COPY internal/ internal/
RUN go mod download

COPY . .
RUN go build -o gorond .

# --- Runtime stage ---
FROM debian:bookworm-slim

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*

RUN mkdir -p /var/log/gorond /var/pid /etc/goron.d

COPY --from=builder /build/gorond /usr/local/bin/gorond

EXPOSE 6777

CMD ["/usr/local/bin/gorond", "-c", "/etc/goron.conf", "-d", "/etc/goron.d/", "--stdout"]
