FROM ubuntu:14.04

MAINTAINER Knut Ahlers <knut@ahlers.me>

WORKDIR /opt

RUN apt-get update && apt-get install -y unzip wget && \
    wget https://gobuilder.me/get/github.com/Luzifer/gobuilder/gobuilder_master_linux-amd64.zip && \
    unzip gobuilder_master_linux-amd64.zip

WORKDIR /opt/gobuilder

ENV PORT 3000

EXPOSE 3000

ENTRYPOINT ["/opt/gobuilder/gobuilder"]
