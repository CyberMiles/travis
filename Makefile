GOTOOLS = github.com/Masterminds/glide

all: get_vendor_deps install

get_vendor_deps: tools
	glide install
	# cannot use ctx (type *"gopkg.in/urfave/cli.v1".Context) as type
	# *"github.com/CyberMiles/travis/vendor/github.com/ethereum/go-ethereum/vendor/gopkg.in/urfave/cli.v1".Context ...
	rm -rf vendor/github.com/ethereum/go-ethereum/vendor/gopkg.in/urfave

install:
	go install ./cmd/travis

tools:
	@echo "--> Installing tools"
	go get $(GOTOOLS)

build: get_vendor_deps
	go build -o build/travis ./cmd/travis

docker:
	docker build -t "ywonline/travis:latest" .

push:
	docker push "ywonline/travis:latest"
