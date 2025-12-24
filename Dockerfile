FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git

WORKDIR /workspace

# Cache deps
COPY go.mod go.sum ./
RUN go mod download

# Copy sources
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/bin/app ./cmd/api

FROM alpine:3.18
RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /app/bin/app /app/app

EXPOSE 8080

ENV GIN_MODE=release

CMD ["/app/app"]
