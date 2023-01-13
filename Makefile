GIT_VERSION = $(shell git rev-parse --short HEAD)

.PHONY: build
build: dist/agent

dist/agent: $(shell find pkg -type f -name '*.go') $(shell find cmd/agent -type f -name '*.go')
	go build -v --ldflags="-w -X main.Version=$(GIT_VERSION)" -o dist/agent cmd/agent/*.go

clean:
	rm -rf dist