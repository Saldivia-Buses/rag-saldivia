FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.work go.work.sum ./
COPY pkg/ ./pkg/
COPY gen/ ./gen/
COPY services/erp/ ./services/erp/

WORKDIR /app/services/erp
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/erp ./cmd/...

FROM gcr.io/distroless/static-debian12
COPY --from=builder /bin/erp /erp
USER nonroot:nonroot
EXPOSE 8013
ENTRYPOINT ["/erp"]
