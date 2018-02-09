all: get_vendor_deps install test

get_vendor_deps:
	go get github.com/Masterminds/glide
	glide install
	# cannot use ctx (type *"gopkg.in/urfave/cli.v1".Context) as type *"github.com/CyberMiles/travis/vendor/github.com/ethereum/go-ethereum/vendor/gopkg.in/urfave/cli.v1".Context ...
	rm -rf vendor/github.com/ethereum/go-ethereum/vendor/gopkg.in/urfave
install:
	go install ./cmd/travis

