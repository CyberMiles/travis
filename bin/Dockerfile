FROM ubuntu:17.10

ENV LIBENI_SRC=/usr/local/lib
ENV LIBENI_PATH=/root/.travis/lib/eni

SHELL ["/bin/bash", "-c"]
RUN apt update \
	&& apt upgrade -y \
	&& apt install -y wget git \
	&& apt install -y build-essential \
	&& apt install -y bison curl \
	&& apt install -y libboost-all-dev cmake openssl libssl-dev \

	&& wget https://github.com/skymizer/SkyPat/releases/download/v3.1.1/skypat-3.1.1.deb \
	&& dpkg -i skypat-3.1.1.deb \

	&& mkdir -p $LIBENI_PATH \
	&& cd $LIBENI_SRC \
	&& git clone https://github.com/CyberMiles/libeni \
	&& cd libeni \
	&& mkdir build && cd build && cmake .. && make -j8 \
	&& find . -name "*.so" | xargs -I{} cp {} $LIBENI_PATH \

	&& curl -s -S -L https://raw.githubusercontent.com/moovweb/gvm/master/binscripts/gvm-installer -o gvm.sh \
	&& chmod +x gvm.sh \
	&& ./gvm.sh \
	&& source /root/.gvm/scripts/gvm \
	&& gvm install go1.9.2 -B \
	&& gvm use go1.9.2 --default \

	&& git clone https://github.com/CyberMiles/travis.git $GOPATH/src/github.com/CyberMiles/travis \
	&& cd $GOPATH/src/github.com/CyberMiles/travis \
	&& git checkout lity \
	&& make all
