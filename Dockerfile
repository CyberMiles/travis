# build docker image
# > docker build -t travis .
# initialize:
# > docker run --rm -v $HOME/.travis:/travis travis node init --home /travis
# node start:
# > docker run --rm -v $HOME/.travis:/travis -p 46657:46657 -p 8545:8545 travis node start --home /travis

# build stage
FROM golang:1.9.3 AS build-env

# libeni
ENV LIBENI_PATH=/app/lib
RUN mkdir -p libeni \
  && wget https://github.com/CyberMiles/libeni/releases/download/v1.2.0/libeni-1.2.0_ubuntu-16.04.tgz -P libeni \
  && tar zxvf libeni/*.tgz -C libeni \
  && mkdir -p $LIBENI_PATH && cp libeni/*/lib/*.so $LIBENI_PATH

# get travis source code
WORKDIR /go/src/github.com/CyberMiles/travis
# copy travis source code from local
ADD . .

# get travis source code from github, develop branch by default.
# you may use a build argument to target a specific branch/tag, for example:
# > docker build -t travis --build-arg branch=lity .
# comment ADD statement above and uncomment two statements below:
# ARG branch=develop
# RUN git clone -b $branch https://github.com/CyberMiles/travis.git --recursive --depth 1 .

# build travis
RUN ENI_LIB=$LIBENI_PATH make build

# final stage
FROM ubuntu:16.04

WORKDIR /app

# add the binary
COPY --from=build-env /go/src/github.com/CyberMiles/travis/build/travis .
COPY --from=build-env /app/lib/*.so ./lib/

EXPOSE 8545 26656 26657

ENTRYPOINT ["./travis"]
