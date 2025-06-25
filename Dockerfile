FROM golang:1.24-alpine as builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o maxmind-geoip-authz main.go

FROM gcr.io/distroless/static
COPY --from=builder /app/maxmind-geoip-authz /app/maxmind-geoip-authz
COPY config.yaml /app/config.yaml
ENTRYPOINT ["/app/maxmind-geoip-authz", "serve"]
