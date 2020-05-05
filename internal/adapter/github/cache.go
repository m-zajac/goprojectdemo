package github

import (
	"context"
	"errors"
	"fmt"
	"time"

	lru "github.com/hashicorp/golang-lru"
	"github.com/m-zajac/goprojectdemo/internal/app"
)

// CachedClient wraps github client with caching layer.
type CachedClient struct {
	client        app.GithubClient
	projectsCache *lru.Cache
	statsCache    *lru.Cache
	ttl           time.Duration
}

// NewCachedClient creates new CachedClient instance.
func NewCachedClient(client app.GithubClient, size int, ttl time.Duration) (*CachedClient, error) {
	if size <= 0 {
		return nil, errors.New("cache size must be greater than 0")
	}
	projectsCache, err := lru.New(size)
	if err != nil {
		return nil, fmt.Errorf("creating lru cache for projects: %w", err)
	}
	statsCache, err := lru.New(size)
	if err != nil {
		return nil, fmt.Errorf("creating lru cache for stats: %w", err)
	}

	return &CachedClient{
		client:        client,
		projectsCache: projectsCache,
		statsCache:    statsCache,
		ttl:           ttl,
	}, nil
}

// ProjectsByLanguage returns projects by given programming language name.
func (c *CachedClient) ProjectsByLanguage(ctx context.Context, language string, count int) ([]app.Project, error) {
	key := c.projectsCacheKey(language)
	val, ok := c.projectsCache.Get(key)
	if ok {
		entry := val.(projectsCacheEntry)
		if entry.count >= count && entry.created.Add(c.ttl).After(time.Now()) {
			projects := entry.data
			if len(projects) > count {
				projects = projects[:count]
			}
			return projects, nil
		}
	}

	projects, err := c.client.ProjectsByLanguage(ctx, language, count)
	if err != nil {
		return projects, err
	}

	entry := projectsCacheEntry{
		created: time.Now(),
		count:   count,
		data:    projects,
	}
	c.projectsCache.Add(key, entry)

	return projects, nil
}

// StatsByProject returns stats by given github project params.
func (c *CachedClient) StatsByProject(ctx context.Context, name string, owner string) ([]app.ContributorStats, error) {
	key := c.statsCacheKey(name, owner)
	val, ok := c.statsCache.Get(key)
	if ok {
		entry := val.(statsCacheEntry)
		if entry.created.Add(c.ttl).After(time.Now()) {
			return entry.data, nil
		}
	}

	stats, err := c.client.StatsByProject(ctx, name, owner)
	if err != nil {
		return stats, err
	}

	entry := statsCacheEntry{
		created: time.Now(),
		data:    stats,
	}
	c.statsCache.Add(key, entry)

	return stats, nil
}

func (c *CachedClient) projectsCacheKey(language string) string {
	return language
}

func (c *CachedClient) statsCacheKey(name string, owner string) string {
	return name + "/" + owner
}

type projectsCacheEntry struct {
	created time.Time
	count   int
	data    []app.Project
}

type statsCacheEntry struct {
	created time.Time
	data    []app.ContributorStats
}
