# build docker image
# > docker build -t cybermiles/travis-build -f Dockerfile.build.ubuntu .
FROM ubuntu:16.04

RUN apt-get update \
  && apt-get upgrade -y \
  && apt-get install -y wget git curl \
  && apt-get install -y build-essential

RUN set -eux; \
	url="https://dl.google.com/go/go1.10.3.linux-amd64.tar.gz"; \
	wget -O go.tgz "$url"; \
	tar -C /usr/local -xzf go.tgz; \
	rm go.tgz; \
	\
	export PATH="/usr/local/go/bin:$PATH"; \
	go version

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"
WORKDIR $GOPATH
