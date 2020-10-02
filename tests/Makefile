# lingualeo Makefile

PWD := $(shell pwd)
BASE_DIR := $(shell basename $(PWD))
# Keep an existing GOPATH, make a private one if it is undefined
GOPATH_DEFAULT := $(PWD)/.go
export GOPATH ?= $(GOPATH_DEFAULT)
GOBIN_DEFAULT := $(GOPATH)/bin
export GOBIN ?= $(GOBIN_DEFAULT)
export GO111MODULE := on
TESTARGS_DEFAULT := -v
TESTARGS ?= $(TESTARGS_DEFAULT)
HAS_GOLANGCI := $(shell command -v golangci-lint;)
HAS_GOIMPORTS := $(shell command -v goimports;)
HAS_GOLINT := $(shell command -v golint;)

# CTI targets

$(GOBIN):
	echo "create gobin"
	mkdir -p $(GOBIN)

work: $(GOBIN)

check: work fmt vet goimports golangci golint

cache:
	go clean -testcache

aws: check cache
	go test -tags=aws $(TESTARGS) -timeout 0 -p 1  ./aws...

fmt:
	go fmt ./...

goimports:
ifndef HAS_GOIMPORTS
	echo "installing goimports"
	GO111MODULE=off go get golang.org/x/tools/cmd/goimports
endif
	goimports -d $(shell find . -iname "*.go")

golint:
ifndef HAS_GOLINT
	echo "installing golint"
	GO111MODULE=off go get go get -u golang.org/x/lint/golint
endif
	golint ./...

vet:
	go vet ./...

golangci:
ifndef HAS_GOLANGCI
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.26.0
endif
	golangci-lint run ./...

cover: work
	go test $(TESTARGS) -tags=unit -cover -coverpkg=./ ./...

shell:
	$(SHELL) -i

.PHONY: cover work fmt test check
