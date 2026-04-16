FROM golang:1.25-alpine AS builder

ARG VERSION=dev
ARG GIT_SHA=unknown
ARG BUILD_TIME=unknown

WORKDIR /app

# Copy only the modules this service needs
COPY pkg/ ./pkg/
COPY gen/ ./gen/
COPY services/ws/ ./services/ws/

# Generate a minimal go.work with only the modules present in the container
RUN printf 'go 1.25.0\n\nuse (\n\t./pkg\n\t./gen/go\n\t./services/ws\n)\n' > go.work

WORKDIR /app/services/ws
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-s -w \
      -X github.com/Camionerou/rag-saldivia/pkg/build.Version=${VERSION} \
      -X github.com/Camionerou/rag-saldivia/pkg/build.GitSHA=${GIT_SHA} \
      -X github.com/Camionerou/rag-saldivia/pkg/build.BuildTime=${BUILD_TIME}" \
    -o /bin/ws ./cmd/...

FROM gcr.io/distroless/static-debian12
COPY --from=builder /bin/ws /ws
USER nonroot:nonroot
EXPOSE 8002
ENTRYPOINT ["/ws"]
