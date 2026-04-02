FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy workspace files for module resolution
COPY go.work go.work.sum ./
COPY pkg/ ./pkg/
COPY services/auth/ ./services/auth/

WORKDIR /app/services/auth
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/auth ./cmd/...

FROM gcr.io/distroless/static-debian12
COPY --from=builder /bin/auth /auth
USER nonroot:nonroot
EXPOSE 8001
ENTRYPOINT ["/auth"]
