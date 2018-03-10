GOTOOLS :=	github.com/mitchellh/gox \
			github.com/Masterminds/glide \
			github.com/mattn/go-sqlite3

BUILD_TAGS? := travis

all: get_vendor_deps install test

get_vendor_deps: tools
	glide install
	# cannot use ctx (type *"gopkg.in/urfave/cli.v1".Context) as type *"github.com/CyberMiles/travis/vendor/github.com/ethereum/go-ethereum/vendor/gopkg.in/urfave/cli.v1".Context ...
	rm -rf vendor/github.com/ethereum/go-ethereum/vendor/gopkg.in/urfave

install:
	go install ./cmd/travis

test:
	@echo "--> Running go test"

tools:
	@echo "--> Installing tools"
	go get $(GOTOOLS)
