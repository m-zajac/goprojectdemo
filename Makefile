.PHONY: build start test clean image loadtest proto generate

build:
	go build ./cmd/goprojectdemo
	go build ./cmd/grpcclient

start:
	go build ./cmd/goprojectdemo && ./goprojectdemo

test:
	go test -race ./...

generate: $(shell go env GOPATH)/bin/mockgen
	GOFLAGS="-mod=readonly" go generate ./...

clean:
	rm -f ./goprojectdemo
	rm -f ./grpcclient
	rm -f ./github.data
	rm -f ./Dockerfile

lint: $(shell go env GOPATH)/bin/golint
	@$(shell go env GOPATH)/bin/golint -set_exit_status `go list ./... | grep -v /vendor/`

lint-more: $(shell go env GOPATH)/bin/golangci-lint
	@$(shell go env GOPATH)/bin/golangci-lint run ./...

image:
	ln -sf ./build/Dockerfile .
	docker build -t goprojectdemo .
	rm ./Dockerfile

loadtest:
	wrk --latency -d 15m -t 2 -c 15 -s scripts/loadtest.lua http://localhost:8080

proto: $(shell go env GOPATH)/bin/protoc-gen-go
	protoc -I api api/service.proto --go_out=plugins=grpc:internal/api/grpc

$(shell go env GOPATH)/bin/mockgen:
	GOFLAGS="-mod=readonly" go get github.com/golang/mock/mockgen@v1.4.3

$(shell go env GOPATH)/bin/protoc-gen-go:
	GOFLAGS="-mod=readonly" go get github.com/golang/protobuf/protoc-gen-go

$(shell go env GOPATH)/bin/golint:
	@GOFLAGS="-mod=readonly" go get golang.org/x/lint/golint

$(shell go env GOPATH)/bin/golangci-lint:
	@GOFLAGS="-mod=readonly" go get github.com/golangci/golangci-lint/cmd/golangci-lint