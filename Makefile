# Makefile

PWD := $(shell pwd)
BASE_DIR := $(shell basename $(PWD))
# Keep an existing GOPATH, make a private one if it is undefined
GOPATH_DEFAULT 			:= $(PWD)/.go
export GOPATH 			?= $(GOPATH_DEFAULT)
GOBIN_DEFAULT 			:= $(GOPATH)/bin
export GOBIN 			?= $(GOBIN_DEFAULT)
export GO111MODULE 		:= on
TEST_ARGS_DEFAULT 		:= -v -p 1 -timeout 0
TEST_ARGS 				?= $(TEST_ARGS_DEFAULT)
TEST_PROVIDER_ARGS 		?= -v --timeout 0
HAS_GOLANGCI 			:= $(shell command -v golangci-lint;)
HAS_GOIMPORTS 			:= $(shell command -v goimports;)
GOOS            		?= $(shell go env GOOS)
GOARCH          		?= $(shell go env GOARCH)
VERSION         		?= $(shell git describe --tags --abbrev=8 --exact-match 2> /dev/null || \
                           git describe --match=$(git rev-parse --short=8 HEAD) --always --dirty --abbrev=8)

PROVIDER_NAME 			:= terraform-provider-polkadot

AZURE_LDFLAGS   		:= "-w -s -X 'azure.version.ProviderVersion=${VERSION}'"
AZURE_BINARY    		:= "./${PROVIDER_NAME}"
AZURE_CMD_PACKAGE 		:= ./pkg/providers/azure
AZURE_PROVIDER_PATH 	:= "${HOME}/.terraform.d/plugins/polkadot-failover-mechanism/azure/polkadot/${VERSION}/${GOOS}_${GOARCH}"

# CTI targets

$(GOBIN):
	echo "create gobin"
	mkdir -p $(GOBIN)

work: $(GOBIN)

check: work fmt vet goimports golangci

cache:
	go clean -testcache

clean:
	go clean -cache

test-aws: check cache
	go test -tags=aws $(TEST_ARGS) ./tests/aws...

test-gcp: check cache
	go test -tags=gcp $(TEST_ARGS) ./tests/gcp...

test-azure: cache install-azure-provider
	go test -tags=azure $(TEST_ARGS) ./tests/azure...

test-azure-provider: check
	go test $(TEST_ARGS) ./pkg/helpers...
	go test $(TEST_ARGS) ./pkg/providers/azure...

build-azure-provider: test-azure-provider
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build \
	-ldflags $(AZURE_LDFLAGS) \
	-tags=azure -o $(AZURE_BINARY) \
	$(AZURE_CMD_PACKAGE)

install-azure-provider: build-azure-provider
	mkdir -p $(AZURE_PROVIDER_PATH)
	mv -v $(AZURE_BINARY) $(AZURE_PROVIDER_PATH)

test-all: test-aws test-gcp test-azure
test-providers-all: test-azure-provider
install-all: install-azure-provider

fmt:
	go fmt ./...

goimports:
ifndef HAS_GOIMPORTS
	echo "installing goimports"
	GO111MODULE=off go get -u golang.org/x/tools/cmd/goimports
endif
	goimports -d $(shell find . -iname "*.go")

vet:
	go vet ./...

golangci:
ifndef HAS_GOLANGCI
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.26.0
endif
	golangci-lint run ./...

shell:
	$(SHELL) -i

.PHONY: work \
		fmt \
		test \
		check \
		test-aws \
		test-gcp \
		test-azure \
		test-all \
		test-providers-all\
		build-azure-provider\
		install-azure-provider
