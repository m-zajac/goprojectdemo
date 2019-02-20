# goprojectdemo [![Build Status](https://travis-ci.org/m-zajac/goprojectdemo.svg?branch=master)](https://travis-ci.org/m-zajac/goprojectdemo) [![codecov](https://codecov.io/gh/m-zajac/goprojectdemo/branch/master/graph/badge.svg)](https://codecov.io/gh/m-zajac/goprojectdemo) [![Go Report Card](https://goreportcard.com/badge/github.com/m-zajac/goprojectdemo)](https://goreportcard.com/report/github.com/m-zajac/goprojectdemo) [![GoDoc](https://godoc.org/github.com/m-zajac/goprojectdemo?status.svg)](http://godoc.org/github.com/m-zajac/goprojectdemo)

Simple go application providing API for checking best contributors for given language.

# Purpose

This project is intended to showcase how I'd write small to medium sized go application today.

I'm aiming to demonstrate how you can:
- Write readable, testable code in go.
- Create single responsibility components, that are wired together as dependencies.
- Separate application/business logic from other code (Ports and Adapters pattern!).
- Structure project in a way that makes it easy to achieve that separation, show how to make package dependencies.

Check and see whether I'm doing this the right way :) Feel free to leave any comment or create an issue if something can be done better.

# Requirements

- Service which will return most active (by commits) contributors for most popular (by stars) projects in given language from github API.
- Service should expose REST API.
- Service should expose GRPC server.
- Service should be able to serve requests FAST (much faster than github API).

# Assumptions

- Can use some storage or db for persistent cache.
- Persistent cache could not be fast enough...
- API can return "Accepted" status, if data isn't available at some level of cache yet. In this case the same request should be eventually responded with proper data.

# Development

- Build binaries: `make build`
- Start server: `make start`
- Run tests: `make test`
- Create docker image: `make image`
- Generate grpc code: `make proto`

`make build` generates `grpclient` binary for testing grpc server. Use `./grpcclient -h` for more info.

# Things to do/improve

- GRPC equivalent of http status 202 is not implemented.
- Proper handling githubs 422 responses, return 404.
- Rate limiter is very basic, needs some work.
- There is some untested code...
