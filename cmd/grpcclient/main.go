// Package main implements very simple grpc client that can be used for testing goprojectdemo grpc server.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	appGrpc "github.com/m-zajac/goprojectdemo/transport/grpc"
	"google.golang.org/grpc"
)

var (
	serverAddr    = flag.String("s", "localhost:9090", "The server address in the format of host:port")
	language      = flag.String("lang", "go", "Programming language")
	projectsCount = flag.Int("pc", 5, "Projects count")
	count         = flag.Int("c", 10, "Results count")
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
		ProjectsCount: int32(*projectsCount),
		Count:         int32(*count),
	}
	resp, err := client.MostActiveContributors(context.Background(), &req)
	if err != nil {
		log.Fatalf("server response error: %v", err)
	}

	fmt.Print("   Commits | Login\n")
	fmt.Print("------------------------\n")
	for _, s := range resp.Stat {
		fmt.Printf("%10d | %s\n", s.Commits, s.Contributor.Login)
	}
}
