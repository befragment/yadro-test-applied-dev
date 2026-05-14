FROM golang:1.25 AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /service ./cmd/service
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go install github.com/pressly/goose/v3/cmd/goose@latest

FROM gcr.io/distroless/static-debian12:nonroot AS app

WORKDIR /app

COPY --from=builder --chmod=755 /service /service
COPY --from=builder /app/data /app/data

ENTRYPOINT ["/service"]

# FROM alpine:3.20 AS app-debug

# WORKDIR /app

# COPY --from=builder --chmod=755 /service /service
# COPY --from=builder /app/data /app/data

# ENTRYPOINT ["/service"]

FROM alpine:3.20 AS migrator

COPY --from=builder /go/bin/goose /usr/local/bin/goose
COPY migrations /app/migrations
COPY scripts /app/scripts

RUN chmod +x /usr/local/bin/goose /app/scripts/migrate.sh

ENTRYPOINT ["/bin/sh", "/app/scripts/migrate.sh"]