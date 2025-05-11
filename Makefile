SOURCES ?= $(shell find . -name "*.go" -type f)
BINARY_NAME = ipvsctl

all: clean lint build

.PHONY: build
build:
	GOOS=linux GOARCH=arm64 go build -o dist/${BINARY_NAME} ipvsctl.go

lint:
	@for file in ${SOURCES} ;  do \
		golint $$file ; \
	done

.PHONY: test
test:
	@go test -v ./...

.PHONY: cover
cover:
	@go test -coverprofile=cover.out ./...
	@go tool cover -func=cover.out

.PHONY: clean
clean:
	@rm -rf dist/*
	@rm -f cover.out

.PHONY: release
release: 
	goreleaser --snapshot --rm-dist
