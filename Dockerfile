FROM golang:1.9.4-alpine3.7
MAINTAINER Atsushi Nagase<a@ngs.io>

COPY . /go/src/github.com/ngs/ts-dakoku
RUN go install github.com/ngs/ts-dakoku

ENTRYPOINT ["/go/bin/ts-dakoku"]
EXPOSE 8000
