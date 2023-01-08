.PHONY: build build-linux build-macos build-windows
.PHONY: tools proto-gen proto-format test

VERSION=$(shell git describe --tags)
GIT_COMMIT=$(shell git rev-parse HEAD)
GIT_COMMIT_DATE=$(shell git log -n1 --pretty='format:%cd' --date=format:'%Y%m%d')
REPO=github.com/bnb-chain/inscription
IMAGE_NAME=ghcr.io/bnb-chain/inscription
REPO=github.com/bnb-chain/inscription

ldflags = -X $(REPO)/version.AppVersion=$(VERSION) \
          -X $(REPO)/version.GitCommit=$(GIT_COMMIT) \
          -X $(REPO)/version.GitCommitDate=$(GIT_COMMIT_DATE)

tools:
	curl https://get.ignite.com/cli! | bash

proto-gen:
	#ignite generate proto-go
	cd proto && buf generate && cp -r github.com/bnb-chain/bfs/x/* ../x && rm -rf github.com

proto-format:
	buf format -w


build:
	go build -o build/bin/bfsd -ldflags="$(ldflags)" ./cmd/bfsd/main.go

docker-image:
	go mod vendor # temporary, should be removed after open source
	docker build . -t ${IMAGE_NAME}

test:
	go test ./...
