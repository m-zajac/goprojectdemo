// Package main implements very simple grpc client that can be used for testing goprojectdemo grpc server.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"

	appGrpc "github.com/m-zajac/goprojectdemo/transport/grpc"
	"google.golang.org/grpc"
)

var (
	serverAddr = flag.String("s", "localhost:9090", "The server address in the format of host:port")
	language   = flag.String("l", "go", "Programming language")
)

func main() {
	flag.Parse()

	conn, err := grpc.Dial(*serverAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to dial: %v", err)
	}
	defer conn.Close()
	client := appGrpc.NewServiceClient(conn)

	req := appGrpc.Request{
		Language:      *language,
		ProjectsCount: 5,
		Count:         3,
	}
	resp, err := client.MostActiveContributors(context.Background(), &req)
	if err != nil {
		log.Fatalf("server response error: %v", err)
	}

	b, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		log.Fatalf("encoding response to json error: %v", err)
	}

	println(string(b))
}
