# build stage
FROM golang:1.9 AS build-env

ENV LIBENI_SRC=/usr/local/lib/libeni

RUN mkdir -p $LIBENI_SRC && cd $LIBENI_SRC \
    && rm -f libeni.tar.gz && wget http://18.218.62.0/lib/libeni.tar.gz \
    && tar zxvf libeni.tar.gz -C $LIBENI_SRC

WORKDIR /go/src/github.com/CyberMiles/travis
ADD . .

RUN ENI_LIB=$LIBENI_SRC make build

# final stage
FROM ubuntu:16.04

ENV DATA_ROOT /travis

ENV LIBENI_SRC=/usr/local/lib/libeni

WORKDIR /app

# Now just add the binary
COPY --from=build-env /go/src/github.com/CyberMiles/travis/build/travis .
COPY --from=build-env $LIBENI_SRC/*.so $LIBENI_SRC/

VOLUME $DATA_ROOT

EXPOSE 8545
EXPOSE 26656
EXPOSE 26657

ENTRYPOINT ["./travis"]

CMD ["node", "start", "--home", "/travis"]
