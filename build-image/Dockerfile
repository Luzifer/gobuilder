FROM golang:cross

MAINTAINER Knut Ahlers <knut@ahlers.me>

ENV GIT_BRANCH master
VOLUME /artifacts

RUN apt-get update && \
    apt-get install -y openssh-client rsync zip && \
    go get github.com/Luzifer/gobuilder/builddb_creator && \
    go install github.com/Luzifer/gobuilder/builddb_creator

ADD ./builder.sh /usr/bin/builder.sh

RUN mkdir /root/.ssh && \
    echo "Host *\n\tStrictHostKeyChecking no\n" >> ~/.ssh/config && \
    chmod 700 /root/.ssh

ENTRYPOINT ["/bin/bash", "-e", "/usr/bin/builder.sh"]
