GOHOSTOS:=$(shell go env GOHOSTOS)
GOPATH:=$(shell go env GOPATH)
VERSION=$(shell git describe --tags --always)

.PHONY: init
# init env
init:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/google/gnostic/cmd/protoc-gen-openapi@latest
	go install github.com/google/wire/cmd/wire@latest
	go install github.com/goreleaser/goreleaser@latest


.PHONY: run
run:
	go run ./...

.PHONY: build
build:
	CGO_ENABLED=0 go build -ldflags "-X main.Version=$(VERSION)" -tags wireinject -o service ./cmd/main.go ./cmd/wire_gen.go

.PHONY: api
api:
	buf generate

.PHONY: gen
# generate
gen:
	go mod download
	go get github.com/google/wire/cmd/wire@latest
	go generate ./...

.PHONY: lint
lint:
	golangci-lint run --fix

.PHONY: test
test:
	go fmt $(go list ./... | grep -v vendor/ | grep -v api/)
	go vet $(go list ./... | grep -v vendor/ | grep -v api/)
	go test -race $(go list ./... | grep -v vendor/ | grep -v api/)

.PHONY: release
release:
	goreleaser release --snapshot --clean

.PHONY: sub
sub:
	git submodule init


.PHONY: all
# generate all
all:
	make api;
	make gen;
	make lint;