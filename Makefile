.PHONY: build build-linux build-macos build-windows
.PHONY: tools proto-gen proto-format test

VERSION=$(shell git describe --tags --always)
GIT_COMMIT=$(shell git rev-parse HEAD)
GIT_COMMIT_DATE=$(shell git log -n1 --pretty='format:%cd' --date=format:'%Y%m%d')
REPO=github.com/bnb-chain/greenfield
IMAGE_NAME=ghcr.io/bnb-chain/greenfield
REPO=github.com/bnb-chain/greenfield

ldflags = -X $(REPO)/version.AppVersion=$(VERSION) \
          -X $(REPO)/version.GitCommit=$(GIT_COMMIT) \
          -X $(REPO)/version.GitCommitDate=$(GIT_COMMIT_DATE)

format:
	bash scripts/format.sh

tools:
	curl https://get.ignite.com/cli! | bash

proto-gen:
	cd proto && buf generate && cp -r github.com/bnb-chain/greenfield/x/* ../x && rm -rf github.com

proto-format:
	buf format -w

build:
	CGO_CFLAGS="-O -D__BLST_PORTABLE__" CGO_CFLAGS_ALLOW="-O -D__BLST_PORTABLE__" go build -o build/bin/gnfd -ldflags="$(ldflags)" ./cmd/gnfd/main.go

docker-image:
	go mod vendor # temporary, should be removed after open source
	docker build . -t ${IMAGE_NAME}

test:
	go test ./...

e2e_test:
	go run ./e2e/main.go
