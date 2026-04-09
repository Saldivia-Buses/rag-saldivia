FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.work go.work.sum ./
COPY pkg/ ./pkg/
COPY services/bigbrother/ ./services/bigbrother/

WORKDIR /app/services/bigbrother
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/bigbrother ./cmd/
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/healthcheck ./cmd/healthcheck/

# Alpine (NOT distroless) — needs nmap binary
# EXCEPTION: documented vs bible.md (BigBrother needs nmap for network scanning)
FROM alpine:3.22
COPY --from=builder /bin/bigbrother /bigbrother
COPY --from=builder /bin/healthcheck /healthcheck
RUN apk add --no-cache nmap libcap && \
    adduser -D -u 10001 bigbrother && \
    setcap cap_net_raw+ep /bigbrother && \
    apk del libcap
USER bigbrother
EXPOSE 8012
ENTRYPOINT ["/bigbrother"]
