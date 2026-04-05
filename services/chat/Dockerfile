FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy workspace files for module resolution
COPY go.work go.work.sum ./
COPY pkg/ ./pkg/
COPY services/chat/ ./services/chat/

WORKDIR /app/services/chat
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/chat ./cmd/...

FROM gcr.io/distroless/static-debian12
COPY --from=builder /bin/chat /chat
USER nonroot:nonroot
EXPOSE 8003
ENTRYPOINT ["/chat"]
