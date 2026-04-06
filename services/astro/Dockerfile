FROM golang:1.25-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app

COPY go.work go.work.sum ./
COPY pkg/ ./pkg/
COPY services/astro/ ./services/astro/

WORKDIR /app/services/astro
RUN go mod download
RUN CGO_ENABLED=1 GOOS=linux go build -o /bin/astro ./cmd/...

FROM alpine:3.22
RUN apk add --no-cache ca-certificates
COPY --from=builder /bin/astro /astro

RUN mkdir -p /ephe
VOLUME ["/ephe"]

USER nobody:nobody
EXPOSE 8011
ENTRYPOINT ["/astro"]
