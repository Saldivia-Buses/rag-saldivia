FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy workspace files for module resolution
COPY go.work go.work.sum ./
COPY pkg/ ./pkg/
COPY services/platform/ ./services/platform/

WORKDIR /app/services/platform
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/platform ./cmd/...

FROM gcr.io/distroless/static-debian12
COPY --from=builder /bin/platform /platform
USER nonroot:nonroot
EXPOSE 8006
ENTRYPOINT ["/platform"]
