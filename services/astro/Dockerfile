FROM golang:1.25-alpine AS builder

ARG VERSION=dev
ARG GIT_SHA=unknown
ARG BUILD_TIME=unknown

RUN apk add --no-cache gcc musl-dev

WORKDIR /app

COPY go.work go.work.sum ./
COPY pkg/ ./pkg/
COPY services/astro/ ./services/astro/

WORKDIR /app/services/astro
RUN go mod download
RUN CGO_ENABLED=1 GOOS=linux go build \
    -ldflags "-s -w \
      -X github.com/Camionerou/rag-saldivia/pkg/build.Version=${VERSION} \
      -X github.com/Camionerou/rag-saldivia/pkg/build.GitSHA=${GIT_SHA} \
      -X github.com/Camionerou/rag-saldivia/pkg/build.BuildTime=${BUILD_TIME}" \
    -o /bin/astro ./cmd/...

FROM alpine:3.22
RUN apk add --no-cache ca-certificates
COPY --from=builder /bin/astro /astro

RUN mkdir -p /ephe
VOLUME ["/ephe"]

USER nobody:nobody
EXPOSE 8011
ENTRYPOINT ["/astro"]
