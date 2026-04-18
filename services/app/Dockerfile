FROM golang:1.25-alpine AS builder

ARG VERSION=dev
ARG GIT_SHA=unknown
ARG BUILD_TIME=unknown

WORKDIR /app

# Copy only the modules this service needs (mirrors services/erp/Dockerfile
# — pkg/* plus the single service, and a minimal go.work generated in the
# container so the build doesn't drag in other services).
COPY pkg/ ./pkg/
COPY services/app/ ./services/app/

RUN printf 'go 1.25.0\n\nuse (\n\t./pkg\n\t./services/app\n)\n' > go.work

WORKDIR /app/services/app
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-s -w \
      -X github.com/Camionerou/rag-saldivia/pkg/server.Version=${VERSION} \
      -X github.com/Camionerou/rag-saldivia/pkg/server.GitSHA=${GIT_SHA} \
      -X github.com/Camionerou/rag-saldivia/pkg/server.BuildTime=${BUILD_TIME}" \
    -o /bin/app ./cmd/...

FROM gcr.io/distroless/static-debian12
COPY --from=builder /bin/app /app
USER nonroot:nonroot
EXPOSE 8020
# Binary self-probes /health via pkg/server.RunHealthcheckAndExit — no shell
# needed (distroless has none). Compose healthchecks must also use CMD form.
HEALTHCHECK --interval=10s --timeout=3s --start-period=20s --retries=3 \
  CMD ["/app", "--healthcheck"]
ENTRYPOINT ["/app"]
