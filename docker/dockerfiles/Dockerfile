#syntax=docker/dockerfile:1
FROM golang:1.22 AS builder

WORKDIR /app
COPY . .

ENV CGO_ENABLED=0
RUN go build \
        -ldflags="-s -w" \
        -trimpath \
        -o chorus \
        ./cmd/chorus


FROM scratch

COPY --from=builder /app/chorus /
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /etc/passwd /etc/passwd

USER nobody
ENTRYPOINT ["/chorus", "start"]
