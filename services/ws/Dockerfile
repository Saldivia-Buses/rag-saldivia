FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy workspace files for module resolution
COPY go.work go.work.sum ./
COPY pkg/ ./pkg/
COPY services/ws/ ./services/ws/

WORKDIR /app/services/ws
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/ws ./cmd/...

FROM gcr.io/distroless/static-debian12
COPY --from=builder /bin/ws /ws
USER nonroot:nonroot
EXPOSE 8002
ENTRYPOINT ["/ws"]
