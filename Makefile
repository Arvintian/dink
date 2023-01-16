export SHELL := /bin/bash

IMAGE_PREFIX ?= $(strip dink-)
IMAGE_SUFFIX ?= $(strip )
IMAGES ?= bundle dind
REGISTRY ?= arvintian

GIT_VERSION = $(shell git rev-parse --short HEAD)
BUILD_DIR := ./build
ARCH ?= amd64

ifeq ($(ARCH),arm64)
DOCKERFILE ?= Dockerfile.arm64
IMAGE_SUFFIX = $(addsuffix -arm64 $(IMAGE_SUFFIX))
else
DOCKERFILE ?= Dockerfile
endif

.PHONY: build
build: dist/agent

dist/agent: $(shell find pkg -type f -name '*.go') $(shell find cmd/agent -type f -name '*.go')
	CGO_ENABLED=0 GOOS=linux GOARCH=$(ARCH)
	go build -v --ldflags="-w -X main.Version=$(GIT_VERSION)" -o dist/agent cmd/agent/*.go


images: build	
	for image in $(IMAGES); do                                                        \
	  imageName=$(IMAGE_PREFIX)$${image/\//-}$(IMAGE_SUFFIX);                         \
	  docker build -t ${REGISTRY}/$${imageName}:$(GIT_VERSION)                        \
	    -f $(BUILD_DIR)/$${image}/$(DOCKERFILE) .;                                    \
	done

code-gen:
	go mod vendor
	rm -rf pkg/apis/dink/v1beta1/*.deepcopy.go
	rm -rf pkg/generated
	./hack/update-codegen.sh

clean:
	rm -rf dist