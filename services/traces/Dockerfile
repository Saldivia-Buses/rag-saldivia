FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.work go.work.sum ./
COPY pkg/ ./pkg/
COPY services/traces/ ./services/traces/

WORKDIR /app/services/traces
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/traces ./cmd/...

FROM gcr.io/distroless/static-debian12
COPY --from=builder /bin/traces /traces
USER nonroot:nonroot
EXPOSE 8009
ENTRYPOINT ["/traces"]
