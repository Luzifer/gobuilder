FROM golang:latest

MAINTAINER Knut Ahlers <knut@ahlers.me>

ADD . /go/src/github.com/Luzifer/gobuilder

WORKDIR /go/src/github.com/Luzifer/gobuilder

RUN go get github.com/tools/godep && \
    godep restore && \
    go build

ENV PORT 3000

EXPOSE 3000

ENTRYPOINT ["./gobuilder"]
