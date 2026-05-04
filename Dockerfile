FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /zero-tunnel .

FROM alpine:latest

# Install ca-certificates for Cloudflare API calls
RUN apk --no-cache add ca-certificates

COPY --from=builder /zero-tunnel /usr/local/bin/zero-tunnel

ENTRYPOINT ["zero-tunnel"]
