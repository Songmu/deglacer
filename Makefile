VERSION = $(shell godzil show-version)
CURRENT_REVISION = $(shell git rev-parse --short HEAD)
IMAGE_NAME ?= deglacer
BUILD_HASH ?= $(shell git rev-parse --verify HEAD 2> /dev/null || echo unknown)
BUILD_LDFLAGS = "-s -w -X github.com/Songmu/deglacer.revision=$(CURRENT_REVISION) -extldflags \"-static\""
u := $(if $(update),-u)

.PHONY: deps
deps:
	go get ${u} -d
	go mod tidy

.PHONY: devel-deps
devel-deps:
	go install github.com/Songmu/godzil/cmd/godzil@latest
	go install github.com/tcnksm/ghr@latest

.PHONY: test
test:
	go test

.PHONY: build
build:
	go build -ldflags=$(BUILD_LDFLAGS) -o bin/deglacer ./cmd/deglacer

docker.build: deps docker.build-only docker.tag

docker.build-only:
	docker build . -t $(IMAGE_NAME) --build-arg BUILD_VERSION=$(BUILD_VERSION) --build-arg BUILD_HASH=$(BUILD_HASH)

docker.tag:
ifdef BUILD_VERSION
	docker image tag $(IMAGE_NAME) $(IMAGE_NAME):$(BUILD_VERSION)
endif

.PHONY: install
install:
	go install -ldflags=$(BUILD_LDFLAGS) ./cmd/deglacer

.PHONY: release
release: devel-deps
	godzil release

CREDITS: go.sum deps devel-deps
	godzil credits -w

.PHONY: crossbuild
crossbuild: CREDITS
	godzil crossbuild -pv=v$(VERSION) -build-ldflags=$(BUILD_LDFLAGS) \
      -os=linux,darwin -d=./dist/v$(VERSION) ./cmd/*

.PHONY: upload
upload:
	ghr -body="$$(godzil changelog --latest -F markdown)" v$(VERSION) dist/v$(VERSION)
