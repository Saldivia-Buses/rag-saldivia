FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.work go.work.sum ./
COPY pkg/ ./pkg/
COPY gen/ ./gen/
COPY modules/ ./modules/
COPY services/agent/ ./services/agent/

WORKDIR /app/services/agent
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/agent ./cmd/...

FROM gcr.io/distroless/static-debian12
COPY --from=builder /bin/agent /agent
USER nonroot:nonroot
EXPOSE 8004
ENTRYPOINT ["/agent"]
