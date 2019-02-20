package main

import "time"

// Config is the container for app configuration
type Config struct {
	// HTTPServerAddress - listen address for http server
	HTTPServerAddress string `default:"0.0.0.0:8080"`

	// HTTPProfileServerAddress - listen address for profiler http server. If empty, profiler server is disabled
	HTTPProfileServerAddress string `default:""`

	// GRPCServerAddress - listen address for grpc server
	GRPCServerAddress string `default:"0.0.0.0:9090"`

	// ServiceResponseTimeout - timeout for service execution
	ServiceResponseTimeout time.Duration `default:"30s"`

	// GithubAPIAddress - address for rest api with protocol
	GithubAPIAddress string `default:"https://api.github.com"`

	// GithubAPIToken - auth token for rest github api (optional, rate limit is lower without this token)
	GithubAPIToken string `default:""`

	// GithubAPIRateLimit - max frequency for github rest api calls
	GithubAPIRateLimit float64 `default:"0.5"`

	// GithubClientCacheSize - maximum number of elements in cache for each github client method
	GithubClientCacheSize int `default:"10000"`

	// GithubClientCacheTTL - maximum lifetime for github client cache entries
	GithubClientCacheTTL time.Duration `default:"10m"`

	// GithubDBPath - filepath for bolt db data
	GithubDBPath string `default:"./github.data"`

	// GithubDBBucketName - bolt db bucket name
	GithubDBBucketName string `default:"github"`

	// GithubDBDataTTL - maximum lifetime for staled data in db
	GithubDBDataTTL time.Duration `default:"8h"`

	// GithubDBDataRefreshTTL - maximum lifetime for staled data to be queued for refresh
	GithubDBDataRefreshTTL time.Duration `default:"1h"`
}
