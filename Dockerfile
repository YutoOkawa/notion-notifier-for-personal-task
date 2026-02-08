# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /usr/local/bin/notion-notifier cmd/server/main.go

# Run stage
FROM gcr.io/distroless/static-debian12

COPY --from=builder /usr/local/bin/notion-notifier /usr/local/bin/notion-notifier

ENTRYPOINT ["/usr/local/bin/notion-notifier", "-config", "/etc/config/notion-notifier/config.yaml"]
