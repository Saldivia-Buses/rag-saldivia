FROM golang:1.25-alpine AS builder

WORKDIR /build
COPY go.mod go.sum ./
COPY ../../pkg /build/pkg
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /search ./cmd/...

FROM scratch
COPY --from=builder /search /search
USER 65534
EXPOSE 8010
ENTRYPOINT ["/search"]
