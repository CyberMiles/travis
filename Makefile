GOTOOLS = github.com/Masterminds/glide
ENI_LIB?=$(HOME)/.travis/eni/lib
CGO_LDFLAGS = -L$(ENI_LIB) -Wl,-rpath,$(ENI_LIB)
CGO_LDFLAGS_ALLOW = "-I.*"
UNAME = $(shell uname)

all: get_vendor_deps install print_cybermiles_logo

get_vendor_deps: tools
	glide install
	@# cannot use ctx (type *"gopkg.in/urfave/cli.v1".Context) as type
	@# *"github.com/CyberMiles/travis/vendor/github.com/ethereum/go-ethereum/vendor/gopkg.in/urfave/cli.v1".Context ...
	@rm -rf vendor/github.com/ethereum/go-ethereum/vendor/gopkg.in/urfave

install:
	@echo "\n--> Installing the Travis TestNet\n"
ifeq ($(UNAME), Linux)
	CGO_LDFLAGS="$(CGO_LDFLAGS)" CGO_LDFLAGS_ALLOW="$(CGO_LDFLAGS_ALLOW)" go install ./cmd/travis
endif
ifeq ($(UNAME), Darwin)
	CGO_LDFLAGS_ALLOW="$(CGO_LDFLAGS_ALLOW)" go install ./cmd/travis
endif
	@echo "\n\nTravis, the TestNet for CyberMiles (CMT) has successfully installed!"

tools:
	@echo "--> Installing tools"
	go get $(GOTOOLS)
	@echo "--> Tools installed successfully"

build: get_vendor_deps
ifeq ($(UNAME), Linux)
	CGO_LDFLAGS="$(CGO_LDFLAGS)" CGO_LDFLAGS_ALLOW="$(CGO_LDFLAGS_ALLOW)" go build -o build/travis ./cmd/travis
endif
ifeq ($(UNAME), Darwin)
	CGO_LDFLAGS_ALLOW="$(CGO_LDFLAGS_ALLOW)" go build -o build/travis ./cmd/travis
endif

NAME := cybermiles/travis
LATEST := ${NAME}:latest
#GIT_COMMIT := $(shell git rev-parse --short=8 HEAD)
#IMAGE := ${NAME}:${GIT_COMMIT}

docker_image:
	docker build -t ${LATEST} .

push_tag_image:
	docker tag ${LATEST} ${NAME}:${TAG}
	docker push ${NAME}:${TAG}

push_image:
	docker push ${LATEST}

dist:
	docker run --rm -e "BUILD_TAG=${BUILD_TAG}" -v "${CURDIR}/scripts":/scripts --entrypoint /bin/sh -t ${LATEST} /scripts/dist.ubuntu.sh
	docker build -t ${NAME}:centos -f Dockerfile.centos .
	docker run --rm -e "BUILD_TAG=${BUILD_TAG}" -v "${CURDIR}/scripts":/scripts --entrypoint /bin/sh -t ${NAME}:centos /scripts/dist.centos.sh
	rm -rf build/dist && mkdir -p build/dist && mv -f scripts/*.zip build/dist/

print_cybermiles_logo:
	@echo "\n\n"
	@echo "    cmtt         tt                        cmt       tit ii  ll                 "
	@echo "  ttcmttt        tt                        tttt      ttt ii  ll                 "
	@echo " tt              tt                        cmtc     ittt     ll                 "
	@echo "it               tt                        mt t     titt ii  ll                 "
	@echo "tt      tt   cmt tt cmt     ,ttt    cm cmt mt tt   tt tt ii  ll    cmtt    cmtt "
	@echo "tt      itt   ti ttcmtttt  ttitttt  cmtcmt mt tt   tt tt ii  ll  ttttitt  tttiti"
	@echo "tt       tt  tt  tt    tt  tt   tt  tt     mt  ti  tt tt ii  ll  tt   tt  ti    "
	@echo "tt        t; tt  tt    tt  ttcmttt  tt     mt  tt it  tt ii  ll  ttcmttt  itttt "
	@echo "it,       tt t   tt    tt  ttcmtii  ti     mt   t tt  tt ii  ll  ttcmtii    tttt"
	@echo " cmt      tttt   tti   tt  tt       ti     mt   cmt   tt ii  ll  tt           tt"
	@echo "  ttcmttt  ttt   ttcmttt   ttcmttt  ti     mt   itt   tt ii  ll  tttttt   tcmttt"
	@echo "    iiii   tt    cmtcmt       iii   ii     mt    ii   ii ii  ll   ttii    iiii  "
	@echo "           ti                                                                   "
	@echo "          tt                                                                    "
	@echo "        ttt                                                                     "
	@echo "\n\n"
	@echo "Please visit the following URL for technical testnet instructions < https://github.com/CyberMiles/travis/blob/master/README.md >.\n"
	@echo "Visit our website < https://www.cybermiles.io/ >, to learn more about CyberMiles.\n"
