FROM golang:latest

MAINTAINER Knut Ahlers <knut@ahlers.me>

ENV GOPATH /go:/go/src/github.com/Luzifer/gobuilder/Godeps/_workspace

ADD . /go/src/github.com/Luzifer/gobuilder

WORKDIR /go/src/github.com/Luzifer/gobuilder

RUN go build

ENV PORT 3000

EXPOSE 3000

ENTRYPOINT ["./gobuilder"]
