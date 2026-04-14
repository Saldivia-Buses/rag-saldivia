FROM golang:1.25-alpine AS builder

ARG VERSION=dev
ARG GIT_SHA=unknown
ARG BUILD_TIME=unknown

RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# Copy only the modules this service needs
COPY pkg/ ./pkg/
COPY services/astro/ ./services/astro/

# Generate a minimal go.work with only the modules present in the container
RUN printf 'go 1.25.0\n\nuse (\n\t./pkg\n\t./services/astro\n)\n' > go.work

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
