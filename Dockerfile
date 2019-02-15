FROM golang:1.9.4-alpine3.7
MAINTAINER Atsushi Nagase<a@ngs.io>

COPY . /go/src/github.com/ngs/ts-dakoku
RUN go install github.com/ngs/ts-dakoku

RUN apk --update add tzdata && \
    cp /usr/share/zoneinfo/Asia/Tokyo /etc/localtime && \
    apk del tzdata && \
    rm -rf /var/cache/apk/*

ENTRYPOINT ["/go/bin/ts-dakoku"]
EXPOSE 8000
