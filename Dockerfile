# ---- Build stage ----
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o /usr/local/bin/neocode ./cmd/neocode

# ---- Runtime stage ----
FROM alpine:3.20

# git: needed for git tool; nodejs/npm: needed for MCP npx servers
RUN apk add --no-cache git nodejs npm ca-certificates

COPY --from=builder /usr/local/bin/neocode /usr/local/bin/neocode

# Default working directory for projects
WORKDIR /workspace

ENTRYPOINT ["neocode"]
