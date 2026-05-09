FROM golang:1.25 AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /service ./cmd/service

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder --chmod=755 /service /service

ENTRYPOINT ["/service"]