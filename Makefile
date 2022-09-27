GO=go
GOFLAGS=-race
DEV_BIN=bin
COV_PROFILE=cp.out

.DEFAULT_GOAL=build

fmt:
	$(GO) fmt ./...
.PHONY: fmt

vet: fmt
	$(GO) vet ./...
.PHONY: vet

lint: vet
	golint -set_exit_status=1 ./...
.PHONY: lint

test: vet
	$(GO) clean -testcache
	$(GO) test ./... -v
.PHONY: test

install:
	$(GO) install ./...
.PHONY: install

build:
	$(GO) build $(GOFLAGS) github.com/mdm-code/bibx/...
.PHONY: build

cover:
	$(GO) test -coverprofile=$(COV_PROFILE) -covermode=atomic ./...
	$(GO) tool cover -html=$(COV_PROFILE)
.PHONY: cover

clean:
	$(GO) clean github.com/mdm-code/bibx/...
	$(GO) mod tidy
	$(GO) clean -testcache -cache
	rm -f $(COV_PROFILE)
.PHONY: clean
