FROM golang:1.20 AS builder

WORKDIR /go/src/serve/
ADD cmd/ ./cmd/
ADD internal/ ./internal/

RUN go mod init serve
RUN go fmt ./...
RUN go build cmd/serve/main.go

FROM alpine:3.18

RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2
RUN apk update && apk upgrade
RUN apk add gcompat

COPY --from=builder /go/src/serve/main /usr/local/bin/serve

ENTRYPOINT serve -config /etc/config.json
