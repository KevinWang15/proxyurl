FROM golang:1.22 AS builder

WORKDIR /app
COPY go.mod .
COPY main.go .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o proxyurl

FROM gcr.io/distroless/base-debian12:nonroot
WORKDIR /app
COPY --from=builder /app/proxyurl /app/proxyurl

EXPOSE 8080
ENV CONFIG_PATH=/app/config.json
USER nonroot:nonroot
ENTRYPOINT ["/app/proxyurl"]
