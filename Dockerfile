FROM golang:1.20 AS builder

COPY *.go /go/src/serve/
WORKDIR /go/src/serve/

RUN go mod init
RUN go build serve

FROM alpine:3.18

RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2
RUN apk update && apk upgrade
RUN apk add gcompat

COPY --from=builder /go/src/serve/serve /usr/local/bin

ENTRYPOINT serve -config /etc/config.json
