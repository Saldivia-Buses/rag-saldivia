FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy workspace files for module resolution
COPY go.work go.work.sum ./
COPY pkg/ ./pkg/
COPY services/rag/ ./services/rag/

WORKDIR /app/services/rag
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/rag ./cmd/...

FROM gcr.io/distroless/static-debian12
COPY --from=builder /bin/rag /rag
USER nonroot:nonroot
EXPOSE 8004
ENTRYPOINT ["/rag"]
