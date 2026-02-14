SHELL := bash
.SHELLFLAGS := -ec

BINARY_NAME := mcpedia
BUILD_DIR := $(CURDIR)/build
VERSION := $(shell date +%y.%m.%d)
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
EXT := $(if $(findstring windows,$(GOOS)),.exe,)
TARGET := $(BINARY_NAME)-$(VERSION)-$(GOOS)-$(GOARCH)$(EXT)
DOCKER_ALPINE_VERSION ?= 3.23

CGO_ENABLED := 1
LDFLAGS := -s -w
GOFLAGS := -mod=vendor -trimpath

all: dev

dev:
	CGO_ENABLED=$(CGO_ENABLED) go build $(GOFLAGS) -o $(BINARY_NAME)$(EXT) ./cmd/mcpedia
	@echo "Built $(BINARY_NAME) (dev)"

release:
	@rm -rf $(BUILD_DIR)/$(BINARY_NAME)-* || true
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) \
		go build $(GOFLAGS) -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(TARGET) ./cmd/mcpedia
	@ls -sh $(BUILD_DIR)/$(BINARY_NAME)-*

test:
	CGO_ENABLED=$(CGO_ENABLED) go test $(GOFLAGS) -race -count=1 ./...

test-cover:
	CGO_ENABLED=$(CGO_ENABLED) go test $(GOFLAGS) -race -count=1 -coverprofile=cover.out ./...
	go tool cover -func=cover.out

fmt:
	gofmt -l -d $$(find . -name '*.go' -not -path './vendor/*')
	@test -z "$$(gofmt -l $$(find . -name '*.go' -not -path './vendor/*'))" || (echo "gofmt check failed" && exit 1)

vet:
	CGO_ENABLED=$(CGO_ENABLED) go vet $(GOFLAGS) ./...

docker:
	docker build \
		--build-arg DOCKER_ALPINE_VERSION=$(DOCKER_ALPINE_VERSION) \
		--build-arg MCPEDIA_VERSION=$(VERSION) \
		-t $(BINARY_NAME):$(VERSION) \
		-t $(BINARY_NAME):latest .

clean:
	@rm -rf $(BUILD_DIR) $(BINARY_NAME) $(BINARY_NAME).exe cover.out

.PHONY: all dev release test test-cover fmt vet docker clean
