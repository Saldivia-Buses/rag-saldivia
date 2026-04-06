FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy workspace files for module resolution
COPY go.work go.work.sum ./
COPY pkg/ ./pkg/
COPY services/__SERVICE_NAME__/ ./services/__SERVICE_NAME__/

WORKDIR /app/services/__SERVICE_NAME__
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/__SERVICE_NAME__ ./cmd/...

FROM gcr.io/distroless/static-debian12
COPY --from=builder /bin/__SERVICE_NAME__ /__SERVICE_NAME__
USER nonroot:nonroot
ENTRYPOINT ["/__SERVICE_NAME__"]
