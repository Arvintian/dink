GIT_VERSION = $(shell git rev-parse --short HEAD)

.PHONY: build
build: dist/agent

dist/agent:
	go build -v --ldflags="-w -X main.Version=$(GIT_VERSION)" -o dist/agent cmd/agent/*.go

clean:
	rm -rf dist