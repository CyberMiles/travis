# build docker image
# > docker build -t cybermiles/travis .
# initialize:
# > docker run --rm -v $HOME/.travis:/travis cybermiles/travis node init --home /travis
# node start:
# > docker run --rm -v $HOME/.travis:/travis -p 26657:26657 -p 8545:8545 cybermiles/travis node start --home /travis

# build stage
FROM cybermiles/travis-build AS build-env

# libeni
ENV LIBENI_PATH=/app/lib
RUN mkdir -p libeni \
  && wget https://github.com/CyberMiles/libeni/releases/download/v1.3.7/libeni-1.3.7_ubuntu-16.04.tgz -P libeni \
  && tar zxvf libeni/*.tgz -C libeni \
  && mkdir -p $LIBENI_PATH && cp libeni/*/lib/* $LIBENI_PATH

# get travis source code
WORKDIR /go/src/github.com/CyberMiles/travis
# copy travis source code from local
ADD . .

# get travis source code from github, develop branch by default.
# you may use a build argument to target a specific branch/tag, for example:
# > docker build -t cybermiles/travis --build-arg branch=develop .
# comment ADD statement above and uncomment two statements below:
# ARG branch=develop
# RUN git clone -b $branch https://github.com/CyberMiles/travis.git --recursive --depth 1 .

# build travis
RUN ENI_LIB=$LIBENI_PATH make build

# final stage
FROM ubuntu:16.04

RUN apt-get update \
  && apt-get install -y libssl-dev

WORKDIR /app
ENV ENI_LIBRARY_PATH=/app/lib
ENV LD_LIBRARY_PATH=/app/lib

# add the binary
COPY --from=build-env /go/src/github.com/CyberMiles/travis/build/travis .
COPY --from=build-env /app/lib/* $ENI_LIBRARY_PATH/
RUN sha256sum travis > travis.sha256

EXPOSE 8545 26656 26657

ENTRYPOINT ["./travis"]
