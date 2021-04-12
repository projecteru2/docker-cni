.PHONY: binary

REVISION := $(shell git rev-parse HEAD || unknown)
BUILTAT := $(shell date +%Y-%m-%dT%H:%M:%S)
VERSION := $(shell git describe --tags $(shell git rev-list --tags --max-count=1))
GO_LDFLAGS ?= -X main.REVISION=$(REVISION) \
			  -X main.BUILTAT=$(BUILTAT) \
			  -X main.VERSION=$(VERSION)

deps:
	env GO111MODULE=on go mod download
	env GO111MODULE=on go mod vendor

binary:
	CGO_ENABLED=0 go build -ldflags "$(GO_LDFLAGS)" -o docker-cni

build: deps binary
