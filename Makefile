.PHONY: build start test clean image loadtest proto

build:
	@go build ./cmd/goprojectdemo
	@go build ./cmd/grpcclient

start:
	@go build ./cmd/goprojectdemo && ./goprojectdemo

test:
	@go test -v -race ./...

clean:
	@rm -f ./goprojectdemo
	@rm -f ./grpcclient
	@rm -f ./github.data
	@rm -f ./Dockerfile

image:
	@ln -sf ./build/Dockerfile .
	@docker build -t goprojectdemo .
	@rm ./Dockerfile

loadtest:
	@wrk --latency -d 15m -t 2 -c 15 -s scripts/loadtest.lua http://localhost:8080

proto:
	@protoc -I api api/service.proto --go_out=plugins=grpc:transport/grpc
