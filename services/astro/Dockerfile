# libswe.a is prebuilt against glibc + FORTIFY_SOURCE (uses __sprintf_chk,
# fseeko64, etc.). musl does not provide those symbols, so we use Debian.
FROM golang:1.25-bookworm AS builder

ARG VERSION=dev
ARG GIT_SHA=unknown
ARG BUILD_TIME=unknown

WORKDIR /app

COPY pkg/ ./pkg/
COPY services/astro/ ./services/astro/

RUN printf 'go 1.25.0\n\nuse (\n\t./pkg\n\t./services/astro\n)\n' > go.work

WORKDIR /app/services/astro
RUN go mod download
RUN cp libswe.a /usr/lib/
# swephgo's cgo directive is `-lswe`, but libswe needs libm (sincos/sqrt).
RUN CGO_ENABLED=1 GOOS=linux CGO_LDFLAGS="-lm" go build \
    -ldflags "-s -w \
      -X github.com/Camionerou/rag-saldivia/pkg/build.Version=${VERSION} \
      -X github.com/Camionerou/rag-saldivia/pkg/build.GitSHA=${GIT_SHA} \
      -X github.com/Camionerou/rag-saldivia/pkg/build.BuildTime=${BUILD_TIME}" \
    -o /bin/astro ./cmd/...

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*
COPY --from=builder /bin/astro /astro

RUN mkdir -p /ephe
VOLUME ["/ephe"]

USER nobody:nogroup
EXPOSE 8011
ENTRYPOINT ["/astro"]
