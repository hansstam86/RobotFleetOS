# Multi-stage build: one image with all layer binaries. Run one layer per container via CMD.
FROM golang:1.22-alpine AS builder
WORKDIR /build

COPY go.mod ./
COPY go.sum* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o bin/fleet ./cmd/fleet && \
    CGO_ENABLED=0 go build -o bin/area ./cmd/area && \
    CGO_ENABLED=0 go build -o bin/zone ./cmd/zone && \
    CGO_ENABLED=0 go build -o bin/edge ./cmd/edge

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /build/bin ./bin

# Default: run fleet (override in docker-compose or k8s per service)
ENV MESSAGING_URL=""
EXPOSE 8080
CMD ["./bin/fleet"]
