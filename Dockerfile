# Multi-stage build: one image with all binaries. Override CMD in Kubernetes per service.
FROM golang:1.22-alpine AS builder
WORKDIR /build

COPY go.mod ./
COPY go.sum* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o bin/all ./cmd/all && \
    CGO_ENABLED=0 go build -o bin/fleet ./cmd/fleet && \
    CGO_ENABLED=0 go build -o bin/area ./cmd/area && \
    CGO_ENABLED=0 go build -o bin/zone ./cmd/zone && \
    CGO_ENABLED=0 go build -o bin/edge ./cmd/edge && \
    CGO_ENABLED=0 go build -o bin/mes ./cmd/mes && \
    CGO_ENABLED=0 go build -o bin/wms ./cmd/wms && \
    CGO_ENABLED=0 go build -o bin/traceability ./cmd/traceability && \
    CGO_ENABLED=0 go build -o bin/qms ./cmd/qms && \
    CGO_ENABLED=0 go build -o bin/cmms ./cmd/cmms && \
    CGO_ENABLED=0 go build -o bin/plm ./cmd/plm && \
    CGO_ENABLED=0 go build -o bin/erp ./cmd/erp

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /build/bin ./bin

ENV MESSAGING_URL=""
EXPOSE 8080 8081 8082 8083 8084 8085 8086 8087
CMD ["./bin/all"]
