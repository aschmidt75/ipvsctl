VERSION := `cat VERSION`
SOURCES ?= $(shell find . -name "*.go" -type f)
BINARY_NAME = ipvsctl

all: clean lint build

.PHONY: build
build:
	GOOS=linux GOARCH=amd64 go build -o release/${BINARY_NAME}-linux-amd64 -ldflags="-X main.version=${VERSION}" ipvsctl.go

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
	@rm -rf release/*
	@rm -f cover.out

