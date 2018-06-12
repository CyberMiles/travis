# build stage
FROM golang:1.9 AS build-env

ENV LIBENI_PATH=/app/lib

WORKDIR /root

RUN mkdir -p libeni \
  && wget https://github.com/CyberMiles/libeni/releases/download/v1.1.0/libeni-1.1.0_ubuntu-16.04.tar.gz -P libeni \
  && tar zxvf libeni/*.tar.gz -C libeni \
  && mkdir -p $LIBENI_PATH && cp libeni/*/lib/*.so $LIBENI_PATH

WORKDIR /go/src/github.com/CyberMiles/travis
ADD . .

RUN ENI_LIB=$LIBENI_PATH make build

# final stage
FROM ubuntu:16.04

ENV DATA_ROOT /travis
ENV LIBENI_PATH /app/lib

WORKDIR /app

# Now just add the binary
COPY --from=build-env /go/src/github.com/CyberMiles/travis/build/travis .
COPY --from=build-env $LIBENI_PATH/*.so $LIBENI_PATH/

VOLUME $DATA_ROOT

EXPOSE 8545
EXPOSE 26656
EXPOSE 26657

ENTRYPOINT ["./travis"]

CMD ["node", "start", "--home", "/travis"]
