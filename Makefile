.PHONY: build build-linux build-macos build-windows
.PHONY: tools proto test

VERSION=$(shell git describe --tags)
GIT_COMMIT=$(shell git rev-parse HEAD)
GIT_COMMIT_DATE=$(shell git log -n1 --pretty='format:%cd' --date=format:'%Y%m%d')
REPO=github.com/bnb-chain/greenfield
IMAGE_NAME=ghcr.io/bnb-chain/greenfield
REPO=github.com/bnb-chain/greenfield

ldflags = -X $(REPO)/version.AppVersion=$(VERSION) \
          -X $(REPO)/version.GitCommit=$(GIT_COMMIT) \
          -X $(REPO)/version.GitCommitDate=$(GIT_COMMIT_DATE)

tools:
	curl https://get.ignite.com/cli! | bash

proto:
	ignite generate proto-go 

build: 
	go build -o build/bin/gnfd -ldflags="$(ldflags)" ./cmd/gnfd/main.go

docker-image:
	go mod vendor # temporary, should be removed after open source
	docker build . -t ${IMAGE_NAME}

test:
	go test ./...