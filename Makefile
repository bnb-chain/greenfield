.PHONY: build build-linux build-macos build-windows
.PHONY: tools proto-gen proto-format test e2e_test ci lint

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
	go install github.com/cosmos/gogoproto/protoc-gen-gocosmos
	go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
	go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger

proto-gen:
	cd proto && buf generate && cp -r github.com/bnb-chain/greenfield/x/* ../x  && cp -r github.com/bnb-chain/greenfield/types/* ../types  && rm -rf github.com && go mod tidy

proto-swagger-gen:
	bash ./scripts/protoc-swagger-gen.sh

proto-swagger-check:
	npm install -g swagger-combine
	bash -x ./scripts/protoc-swagger-gen.sh
	git diff --exit-code

proto-format:
	buf format -w

proto-format-check:
	buf format --diff --exit-code

build:
	go build -o build/bin/gnfd -ldflags="$(ldflags)" ./cmd/gnfd/main.go

mock-gen:
	sh ./scripts/mockgen.sh

docker-image:
	go mod vendor # temporary, should be removed after open source
	docker build . -t ${IMAGE_NAME}

test:
	go test -failfast $$(go list ./... | grep -v e2e | grep -v sdk)

e2e_start_localchain:
	bash ./deployment/localup/localup.sh all 1 7

e2e_test:
	go test -p 1 -failfast -v ./e2e/... -timeout 99999s

lint:
	golangci-lint run --fix

# ensures the changes of proto files generate new go files
proto-gen-check: proto-gen
	git diff --exit-code

ci: proto-format-check build test e2e_start_localchain e2e_test lint proto-gen-check
	echo "ci passed"
