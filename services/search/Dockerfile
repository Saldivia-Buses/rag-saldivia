FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy workspace files for module resolution
COPY go.work go.work.sum ./
COPY pkg/ ./pkg/
COPY gen/ ./gen/
COPY services/search/ ./services/search/

WORKDIR /app/services/search
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/search ./cmd/...

FROM gcr.io/distroless/static-debian12
COPY --from=builder /bin/search /search
USER nonroot:nonroot
EXPOSE 8010
ENTRYPOINT ["/search"]
