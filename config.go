package main

import "time"

// Config is the container for app configuration
type Config struct {
	// HTTPServerAddress - listen address for http server
	HTTPServerAddress string `default:"0.0.0.0:8080"`

	// HTTPProfileServerAddress - listen address for profiler http server. If empty, profiler server is disabled
	HTTPProfileServerAddress string `default:""`

	// GithubAPIAddress - address for rest api with protocol
	GithubAPIAddress string `default:"https://api.github.com"`

	// GithubAPIToken - auth token for rest github api (optional, rate limit is lower without this token)
	GithubAPIToken string `default:""`

	// GithubAPIRateLimit - max frequency for github rest api calls
	GithubAPIRateLimit int `default:"5"`

	// GithubTimeout - timeout for github api calls
	GithubTimeout time.Duration `default:"15s"`

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
}
