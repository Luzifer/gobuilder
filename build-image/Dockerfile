FROM golang:cross

MAINTAINER Knut Ahlers <knut@ahlers.me>

VOLUME /artifacts

RUN apt-get update && \
    apt-get install -y openssh-client rsync zip wget gnupg && \
    go get github.com/Luzifer/gobuilder/cmd/configreader

ADD ./builder.sh /usr/bin/builder.sh
ADD ./gpgkey.asc.enc /root/gpgkey.asc.enc

RUN mkdir /root/.ssh && \
    echo "Host *\n\tStrictHostKeyChecking no\n" >> ~/.ssh/config && \
    chmod 700 /root/.ssh && \
    gpg --list-keys 2>&1 1>/dev/null && \
    echo "keyserver-options auto-key-retrieve" >> ~/.gnupg/gpg.conf && \
    sed -i "s/^keyserver .*$/keyserver hkp:\/\/keyserver.ubuntu.com/" ~/.gnupg/gpg.conf

ENTRYPOINT ["/bin/bash", "-e", "/usr/bin/builder.sh"]
