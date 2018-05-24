# build stage
FROM golang:1.9 AS build-env

ENV LIBENI_SRC=/usr/local/lib
ENV LIBENI_PATH=/usr/local/lib/libeni

RUN apt-get update \
  && apt-get upgrade -y \
  && apt-get install -y bison libboost-all-dev cmake libssl-dev

WORKDIR /root

RUN wget https://github.com/skymizer/SkyPat/releases/download/v3.1.1/skypat-3.1.1.deb \
  && dpkg -i skypat-3.1.1.deb

RUN mkdir -p $LIBENI_PATH \
  && cd $LIBENI_SRC \
  && git clone https://github.com/CyberMiles/libeni \
  && cd libeni \
  && mkdir build && cd build && cmake .. && make -j8 \
  && find . -name "*.so" | xargs -I{} cp {} $LIBENI_PATH

WORKDIR /go/src/github.com/CyberMiles/travis
ADD . .

RUN ENI_LIB=$LIBENI_PATH make build

# final stage
FROM ubuntu:16.04

ENV DATA_ROOT /travis

ENV LIBENI_PATH=/usr/local/lib/libeni

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
