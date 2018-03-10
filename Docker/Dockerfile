FROM golang:1.9-alpine

ENV DATA_ROOT /travis
ENV TRHOME $DATA_ROOT

RUN addgroup tuser && \
    adduser -S -G tuser tuser

RUN mkdir -p $DATA_ROOT && \
    chown -R tuser:tuser $DATA_ROOT

RUN apk add --no-cache bash curl jq

RUN mkdir - /root/.glide
ADD ./mirrors.yaml /root/.glide/

ENV GOPATH /go
ENV PATH "$PATH:/go/bin"
RUN mkdir -p /go/src/github.com/CyberMiles/travis && \
    apk add --no-cache build-base git linux-headers && \
    cd /go/src/github.com/CyberMiles/travis && \
    git clone https://github.com/CyberMiles/travis . && \
    git checkout deploy && \
    make get_vendor_deps && \
    NOERROR=1 && make install && \
    glide cc && \
    cd - && \
    rm -rf /go/src/github.com/CyberMiles/travis && \
    apk del build-base git linux-headers

VOLUME $DATA_ROOT

EXPOSE 8545
EXPOSE 46656
EXPOSE 46657

ENTRYPOINT ["travis"]

CMD ["node", "start", "--home=${TRHOME}"]
