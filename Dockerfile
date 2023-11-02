FROM golang:1.20-alpine AS builder

RUN go env -w GO111MODULE=on \
  && go env -w CGO_ENABLED=0 \
  && go env

RUN apk update && apk add git

WORKDIR /build

COPY ./ .

RUN set -ex \
    && go mod tidy \
    && go build -ldflags "-s -w" -o zerobot -trimpath

FROM alpine:latest

RUN apk add --no-cache ffmpeg

COPY --from=builder /build/zerobot /usr/bin/zerobot
RUN chmod +x /usr/bin/zerobot

WORKDIR /data

ENTRYPOINT [ "/usr/bin/zerobot" ]
