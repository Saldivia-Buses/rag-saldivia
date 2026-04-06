FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.work go.work.sum ./
COPY pkg/ ./pkg/
COPY services/feedback/ ./services/feedback/

WORKDIR /app/services/feedback
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/feedback ./cmd/...

FROM gcr.io/distroless/static-debian12
COPY --from=builder /bin/feedback /feedback
USER nonroot:nonroot
EXPOSE 8008
ENTRYPOINT ["/feedback"]
