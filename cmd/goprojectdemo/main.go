package main

import (
	netHttp "net/http"
	"sync"
	"time"

	"github.com/m-zajac/goprojectdemo/database"
	"github.com/m-zajac/goprojectdemo/limiter"
	"github.com/sirupsen/logrus"

	"github.com/m-zajac/goprojectdemo/app"

	"github.com/kelseyhightower/envconfig"
	"github.com/m-zajac/goprojectdemo/adapter/github"
	"github.com/m-zajac/goprojectdemo/transport/grpc"
	"github.com/m-zajac/goprojectdemo/transport/http"
)

func main() {
	l := logrus.New()
	l.Level = logrus.InfoLevel

	var conf Config
	if err := envconfig.Process("", &conf); err != nil {
		l.Fatalf("coludn't parse config: %v", err)
	}

	httpClient := &netHttp.Client{
		Timeout: 30 * time.Second,
	}
	limitedHTTPClient := limiter.NewLimitedHTTPDoer(
		httpClient,
		conf.GithubAPIRateLimit,
	)

	kvStore, err := database.NewBoltKVStore(
		conf.GithubDBPath,
		conf.GithubDBBucketName,
	)
	if err != nil {
		l.Fatalf("coludn't create bolt kv store: %v", err)
	}
	defer kvStore.Close()

	githubClient := github.NewClient(
		limitedHTTPClient,
		conf.GithubAPIAddress,
		conf.GithubAPIToken,
		conf.GithubTimeout,
	)
	githubStaleDataClient, err := github.NewClientWithStaleData(
		githubClient,
		kvStore,
		conf.GithubDBDataTTL,
		l.WithField("component", "githubStaleDataClient"),
	)
	if err != nil {
		l.Fatalf("coludn't create github db client: %v", err)
	}
	githubStaleDataClient.RunScheduler()
	defer githubStaleDataClient.Close()
	githubCachedClient, err := github.NewCachedClient(
		githubStaleDataClient,
		conf.GithubClientCacheSize,
		conf.GithubClientCacheTTL,
	)
	if err != nil {
		l.Fatalf("couldn't create github client cache: %v", err)
	}

	service := app.NewService(
		githubCachedClient,
	)

	mux := http.NewMux(service, 60*time.Second, l.WithField("component", "mux"))
	server := http.NewServer(
		conf.HTTPServerAddress,
		conf.HTTPProfileServerAddress,
		mux,
		l.WithField("component", "httpServer"),
	)

	grpcService := grpc.NewService(service)
	grpcServer := grpc.NewServer(
		grpcService,
		conf.GRPCServerAddress,
		l.WithField("component", "grpcServer"),
	)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		server.Run()
		wg.Done()
	}()
	wg.Add(1)
	go func() {
		if err := grpcServer.Run(); err != nil {
			l.Fatalf("couldn't run grpc server: %v", err)
		}
		wg.Done()
	}()
	wg.Wait()
}
