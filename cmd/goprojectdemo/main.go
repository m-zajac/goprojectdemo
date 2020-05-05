package main

import (
	netHttp "net/http"
	"sync"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/m-zajac/goprojectdemo/internal/adapter/github"
	"github.com/m-zajac/goprojectdemo/internal/api/grpc"
	"github.com/m-zajac/goprojectdemo/internal/api/http"
	"github.com/m-zajac/goprojectdemo/internal/api/http/limiter"
	"github.com/m-zajac/goprojectdemo/internal/app"
	"github.com/m-zajac/goprojectdemo/internal/database"
	"github.com/sirupsen/logrus"
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
	limitedHTTPClient := limiter.NewHTTPDoer(
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
	)
	githubStaleDataClient, err := github.NewClientWithStaleData(
		githubClient,
		kvStore,
		conf.GithubDBDataTTL,
		conf.GithubDBDataRefreshTTL,
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
		conf.ServiceResponseTimeout,
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
