# build stage
FROM golang:1.9-alpine3.6 AS build-env

RUN apk add --no-cache bash build-base curl jq git linux-headers

WORKDIR /go/src/github.com/CyberMiles/travis
ADD . .

RUN make build

# final stage
FROM alpine:3.6

ENV DATA_ROOT /travis

WORKDIR /app

# Now just add the binary
COPY --from=build-env /go/src/github.com/CyberMiles/travis/build/travis .

VOLUME $DATA_ROOT

EXPOSE 8545
EXPOSE 26656
EXPOSE 26657

ENTRYPOINT ["./travis"]

CMD ["node", "start", "--home", "/travis"]
