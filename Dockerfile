# build stage
FROM golang:1.9 AS build-env

ENV LIBENI_PATH=/app/lib

WORKDIR /root

RUN mkdir -p libeni \
  && wget https://github.com/CyberMiles/libeni/releases/download/v1.0.x/libeni.tar.gz -P libeni \
  && tar zxvf libeni/libeni.tar.gz \
  && mkdir -p $LIBENI_PATH && cp libeni/ubuntu/eni/lib/*.so $LIBENI_PATH

WORKDIR /go/src/github.com/CyberMiles/travis
ADD . .

RUN ENI_LIB=$LIBENI_PATH make build

# final stage
FROM ubuntu:16.04

ENV DATA_ROOT /travis
ENV LIBENI_PATH=/app/lib

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
