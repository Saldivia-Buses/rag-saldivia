FROM golang:1.25-alpine AS builder

ARG VERSION=dev
ARG GIT_SHA=unknown
ARG BUILD_TIME=unknown

WORKDIR /app

# Copy only the modules this service needs
COPY pkg/ ./pkg/
COPY services/feedback/ ./services/feedback/

# Generate a minimal go.work with only the modules present in the container
RUN printf 'go 1.25.0\n\nuse (\n\t./pkg\n\t./services/feedback\n)\n' > go.work

WORKDIR /app/services/feedback
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-s -w \
      -X github.com/Camionerou/rag-saldivia/pkg/build.Version=${VERSION} \
      -X github.com/Camionerou/rag-saldivia/pkg/build.GitSHA=${GIT_SHA} \
      -X github.com/Camionerou/rag-saldivia/pkg/build.BuildTime=${BUILD_TIME}" \
    -o /bin/feedback ./cmd/...

FROM gcr.io/distroless/static-debian12
COPY --from=builder /bin/feedback /feedback
USER nonroot:nonroot
EXPOSE 8008
ENTRYPOINT ["/feedback"]
