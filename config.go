package main

import "time"

// Config is the container for app configuration
type Config struct {
	Address          string        `default:"0.0.0.0:8080"`
	GithubAPIAddress string        `default:"https://api.github.com"`
	GithubAPIToken   string        `default:""`
	GithubTimeout    time.Duration `default:"10s"`
	CacheSize        int           `default:"10000"`
	CacheTTL         time.Duration `default:"10m"`
}
