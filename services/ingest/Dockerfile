FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy workspace files for module resolution
COPY go.work go.work.sum ./
COPY pkg/ ./pkg/
COPY services/ingest/ ./services/ingest/

WORKDIR /app/services/ingest
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/ingest ./cmd/...

FROM gcr.io/distroless/static-debian12
COPY --from=builder /bin/ingest /ingest
USER nonroot:nonroot
EXPOSE 8007
ENTRYPOINT ["/ingest"]
