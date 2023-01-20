export SHELL := /bin/bash

IMAGE_PREFIX ?= $(strip dink-)
IMAGE_SUFFIX ?= $(strip )
IMAGES ?= bundle dind
REGISTRY ?= arvintian

GIT_VERSION = $(shell git rev-parse --short HEAD)
BINS = agent controller server play
ARCH ?= amd64

BUILD_DIR := build
ifeq ($(ARCH),arm64)
DOCKERFILE ?= Dockerfile.arm64
IMAGE_SUFFIX = $(addsuffix -arm64 $(IMAGE_SUFFIX))
else
DOCKERFILE ?= Dockerfile
endif

.PHONY: build
build: $(shell find pkg -type f -name '*.go') $(shell find cmd -type f -name '*.go')
	for bin in $(BINS); do \
		CGO_ENABLED=0 GOOS=linux GOARCH=$(ARCH) go build -v --ldflags="-w -X main.Version=$(GIT_VERSION)" -o dist/$${bin} cmd/$${bin}/*.go; \
	done

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