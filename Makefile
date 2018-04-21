GOTOOLS = github.com/Masterminds/glide

all: get_vendor_deps install

get_vendor_deps: tools
	glide install
	@# cannot use ctx (type *"gopkg.in/urfave/cli.v1".Context) as type
	@# *"github.com/CyberMiles/travis/vendor/github.com/ethereum/go-ethereum/vendor/gopkg.in/urfave/cli.v1".Context ...
	@rm -rf vendor/github.com/ethereum/go-ethereum/vendor/gopkg.in/urfave

install:
	@echo "--> Installing the Travis TestNet"
	go install ./cmd/travis
	@echo "\nTravis TestNet installed successfully\n"
	@echo "Please visit the following URL for further instructions on initializing and running the TestNet < https://github.com/CyberMiles/travis/blob/master/README.md >.\n"

tools:
	@echo "--> Installing tools"
	go get $(GOTOOLS)
	@echo "--> Tools installed successfully"

build: get_vendor_deps
	go build -o build/travis ./cmd/travis

IMAGE := ywonline/travis

docker_image:
	docker build -t $(IMAGE) .

push_image:
	docker push $(IMAGE)
