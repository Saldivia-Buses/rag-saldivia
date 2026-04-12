FROM golang:1.25-alpine AS builder

ARG VERSION=dev
ARG GIT_SHA=unknown
ARG BUILD_TIME=unknown

WORKDIR /app

COPY go.work go.work.sum ./
COPY pkg/ ./pkg/
COPY gen/ ./gen/
COPY modules/ ./modules/
COPY services/agent/ ./services/agent/

WORKDIR /app/services/agent
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-s -w \
      -X github.com/Camionerou/rag-saldivia/pkg/build.Version=${VERSION} \
      -X github.com/Camionerou/rag-saldivia/pkg/build.GitSHA=${GIT_SHA} \
      -X github.com/Camionerou/rag-saldivia/pkg/build.BuildTime=${BUILD_TIME}" \
    -o /bin/agent ./cmd/...

FROM gcr.io/distroless/static-debian12
COPY --from=builder /bin/agent /agent
COPY --from=builder /app/modules /modules
USER nonroot:nonroot
EXPOSE 8004
ENTRYPOINT ["/agent"]
