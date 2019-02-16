package main

import (
	"log"
	netHttp "net/http"
	"time"

	"github.com/m-zajac/goprojectdemo/app"

	"github.com/kelseyhightower/envconfig"
	"github.com/m-zajac/goprojectdemo/adapters/github"
	"github.com/m-zajac/goprojectdemo/transport/http"
)

func main() {
	var conf Config
	if err := envconfig.Process("", &conf); err != nil {
		log.Fatalf("coludn't parse config: %v", err.Error())
	}

	httpClient := &netHttp.Client{
		Timeout: 30 * time.Second,
	}

	githubClient := github.NewClient(
		httpClient,
		conf.GithubAPIAddress,
		conf.GithubAPIToken,
		conf.GithubTimeout,
	)

	service := app.NewService(
		githubClient,
	)

	mux := http.NewMux(service, 60*time.Second)
	server := http.NewServer(
		conf.Address,
		mux,
	)

	server.Run()
}
