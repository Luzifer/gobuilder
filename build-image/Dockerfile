FROM golang:1.7

MAINTAINER Knut Ahlers <knut@ahlers.me>

VOLUME /artifacts

RUN set -ex \
 && go version \
 && apt-get update \
 && apt-get install -y openssh-client rsync zip wget gnupg \
 && mkdir -p /go/src/github.com/Luzifer \
 && git clone https://github.com/Luzifer/gobuilder.git /go/src/github.com/Luzifer/gobuilder \
 && go install github.com/Luzifer/gobuilder/cmd/configreader \
 && go install github.com/Luzifer/gobuilder/cmd/asset-sync \
 && rm -rf /go/src/*

ADD ./builder.sh /usr/bin/builder.sh
ADD ./gpgkey.asc.enc /root/gpgkey.asc.enc

RUN mkdir /root/.ssh \
 && echo "Host *\n\tStrictHostKeyChecking no\n" >> ~/.ssh/config \
 && chmod 700 /root/.ssh \
 && gpg --list-keys 2>&1 1>/dev/null \
 && echo "keyserver-options auto-key-retrieve" >> ~/.gnupg/gpg.conf \
 && sed -i "s/^keyserver .*$/keyserver hkp:\/\/keyserver.ubuntu.com/" ~/.gnupg/gpg.conf

ENTRYPOINT ["/bin/bash", "-e", "/usr/bin/builder.sh"]
