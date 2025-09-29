GOHOSTOS:=$(shell go env GOHOSTOS)
GOPATH:=$(shell go env GOPATH)
VERSION=$(shell git describe --tags --always)
APP_NAME=sweets-app
CONFIG_FILE=etc/config.yaml

.PHONY: init
# init env - install required tools
init:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/google/wire/cmd/wire@latest
	go install github.com/cloudwego/kitex/tool/cmd/kitex@latest
	go install github.com/cloudwego/thriftgo@latest
	go install github.com/pressly/goose/v3/cmd/goose@latest
	go install github.com/bufbuild/buf/cmd/buf@latest
	go install github.com/envoyproxy/protoc-gen-validate@latest

.PHONY: run
# run service in integrated mode (HTTP + RPC)
run:
	go run ./cmd/... -f $(CONFIG_FILE)

.PHONY: build
# build binary
build:
	CGO_ENABLED=0 go build -ldflags "-X main.Version=$(VERSION)" -o $(APP_NAME) ./cmd/...

.PHONY: proto-buf
# generate protobuf code with buf
proto-buf:
	cd api && buf generate

.PHONY: proto-kitex
# generate Kitex service code
proto-kitex:
	kitex -module github.com/go-sweets/sweets-layout \
	      -service hello \
	      -I ./api/proto \
	      -gen-path ./api/gen/kitex \
	      -no-main \
	      ./api/proto/hello_kitex.proto

.PHONY: proto
# generate all proto code (buf + kitex)
proto: proto-buf proto-kitex

.PHONY: proto-clean
# clean generated proto code
proto-clean:
	rm -rf api/gen/

.PHONY: wire
# generate wire dependency injection
wire:
	cd cmd && wire

.PHONY: gen
# generate all code (proto + wire)
gen: proto wire

.PHONY: fmt
# format code
fmt:
	go fmt ./...
	goimports -w .

.PHONY: lint
# run linter
lint:
	golangci-lint run --fix

.PHONY: test
# run tests with race detection
test:
	go test -race -v ./...

.PHONY: test-coverage
# run tests with coverage
test-coverage:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

.PHONY: migrate-up
# run database migrations
migrate-up:
	cd internal/db && goose mysql "$(shell grep DSN etc/config.yaml | cut -d'"' -f2)" up

.PHONY: migrate-down
# rollback last migration
migrate-down:
	cd internal/db && goose mysql "$(shell grep DSN etc/config.yaml | cut -d'"' -f2)" down

.PHONY: migrate-status
# check migration status
migrate-status:
	cd internal/db && goose mysql "$(shell grep DSN etc/config.yaml | cut -d'"' -f2)" status

.PHONY: migrate-create
# create new migration: make migrate-create name=create_users_table
migrate-create:
	cd internal/db && goose create $(name) sql

.PHONY: docker-build
# build docker image
docker-build:
	docker build -t $(APP_NAME):$(VERSION) .

.PHONY: docker-run
# run in docker
docker-run:
	docker run -p 8080:8080 -p 9090:9090 -v $(PWD)/etc:/app/etc $(APP_NAME):$(VERSION)

.PHONY: clean
# clean build artifacts
clean:
	rm -f $(APP_NAME)
	rm -f coverage.out coverage.html
	find . -name "*.test" -type f -delete
	find . -name "*.backup" -type f -delete

.PHONY: dev
# run in development mode with hot reload
dev:
	air -c .air.toml

.PHONY: all
# run all generation and checks
all: gen fmt lint test

.PHONY: help
# show help
help:
	@echo ''
	@echo 'Usage:'
	@echo ' make [target]'
	@echo ''
	@echo 'Targets:'
	@awk '/^[a-zA-Z\-\_0-9]+:/ { \
	helpMessage = match(lastLine, /^# (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")-1); \
			helpMessage = substr(lastLine, RSTART + 2, RLENGTH); \
			printf "\033[36m%-20s\033[0m %s\n", helpCommand,helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help