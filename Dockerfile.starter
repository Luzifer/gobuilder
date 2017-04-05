FROM golang

LABEL maintainer "Knut Ahlers <knut@ahlers.me>"

ADD . /go/src/github.com/Luzifer/gobuilder
WORKDIR /go/src/github.com/Luzifer/gobuilder

RUN set -ex \
 && apt-get update \
 && apt-get install -y git ca-certificates \
 && go install -ldflags "-X main.version=$(git describe --tags || git rev-parse --short HEAD || echo dev)" github.com/Luzifer/gobuilder/cmd/starter \
 && apt-get remove -y --purge git \
 && apt-get autoremove -y

VOLUME /data/gobuilder-starter
VOLUME /var/run/docker.sock
VOLUME /root

ENTRYPOINT ["/go/bin/starter"]
CMD ["--"]
