.PHONY: docs
REPO:=github.com/CyberMiles/travis/sdk/go-wire/data

docs:
	@go get github.com/davecheney/godoc2md
	godoc2md $(REPO) > README.md

test:
	go test ./...
