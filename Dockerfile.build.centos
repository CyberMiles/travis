# build docker image
# > docker build -t cybermiles/travis-build:centos -f Dockerfile.build.centos .
FROM centos:7

RUN yum update -y \
  && yum install -y wget git curl openssl-devel \
  && yum group install -y "Development Tools"

# install go
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
