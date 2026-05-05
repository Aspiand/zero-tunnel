FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION="(dev)"
ARG COMMIT="(dev)"
ARG DATE="(dev)"

RUN go build -ldflags "-s -w -X main.Version=${VERSION} -X main.Commit=${COMMIT} -X main.BuildDate=${DATE}" -o /zero-tunnel .

FROM alpine:latest

# Install ca-certificates for Cloudflare API calls
RUN apk --no-cache add ca-certificates

COPY --from=builder /zero-tunnel /usr/local/bin/zero-tunnel

ENTRYPOINT ["zero-tunnel"]
