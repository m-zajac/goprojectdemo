.PHONY: start test clean image loadtest

start:
	@go build && ./goprojectdemo 

test:
	@go test -v -race ./...

clean:
	@rm -f ./goprojectdemo
	@rm -f ./github.data
	@rm -f ./Dockerfile

image:
	@ln -sf ./build/Dockerfile .
	@docker build -t goprojectdemo .
	@rm ./Dockerfile

loadtest:
	@wrk --latency -d 15m -s scripts/loadtest.lua http://localhost:8080