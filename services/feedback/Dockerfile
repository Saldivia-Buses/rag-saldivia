FROM golang:1.25 AS builder
WORKDIR /app
COPY go.mod go.sum ./
COPY ../../pkg/ /pkg/
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /feedback ./cmd/

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /feedback /feedback
USER nonroot:nonroot
EXPOSE 8008
ENTRYPOINT ["/feedback"]
