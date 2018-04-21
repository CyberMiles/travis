GOTOOLS = github.com/Masterminds/glide

all: 
        @echo "--> Installing the CyberMiles Travis TestNet"
	get_vendor_deps 
	install
	@echo "--> Installation has completed successfully."

get_vendor_deps: 
        @echo "--> Installing dependencies"
        tools
	glide install
	rm -rf vendor/github.com/ethereum/go-ethereum/vendor/gopkg.in/urfave

install:
	go install ./cmd/travis

tools:
	@echo "--> Installing tools"
	go get $(GOTOOLS)

build: get_vendor_deps
	go build -o build/travis ./cmd/travis

IMAGE := ywonline/travis

docker_image:
	docker build -t $(IMAGE) .

push_image:
	docker push $(IMAGE)
