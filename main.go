package main

import (
	"log"
	netHttp "net/http"
	"time"

	"github.com/m-zajac/goprojectdemo/limiter"

	"github.com/m-zajac/goprojectdemo/app"

	"github.com/kelseyhightower/envconfig"
	"github.com/m-zajac/goprojectdemo/adapter/github"
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
	limitedHTTPClient := limiter.NewLimitedHTTPDoer(
		httpClient,
		conf.GithubAPIRateLimit,
	)

	githubClient := github.NewClient(
		limitedHTTPClient,
		conf.GithubAPIAddress,
		conf.GithubAPIToken,
		conf.GithubTimeout,
	)
	githubCachedClient, err := github.NewCachedClient(
		githubClient,
		conf.GithubClientCacheSize,
		conf.GithubClientCacheTTL,
	)
	if err != nil {
		log.Fatalf("couldn't create github client cache: %v", err)
	}

	service := app.NewService(
		githubCachedClient,
	)

	mux := http.NewMux(service, 60*time.Second)
	server := http.NewServer(
		conf.HTTPServerAddress,
		mux,
	)

	server.Run()
}
