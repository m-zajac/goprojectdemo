package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/m-zajac/goprojectdemo/internal/app"
	"github.com/sirupsen/logrus"
)

// KVStore provides simple kv data storage
type KVStore interface {
	ReadKey(key []byte) ([]byte, error)
	UpdateKey(key []byte, data []byte) error
}

// ClientWithStaleData wraps GithubClient and returns data saved in db if possible.
//
// If data is not available (or datas ttl is exceeded), update is scheduled, and app.ScheduledForLaterError is returned with empty data.
// If data is available, ttl is ok, but refreshTTL is exceeded, additional job for update is scheduled. Exisiting data is returned immediately.
// If data is available and no ttl is exceeded, then data is returned immediately.
type ClientWithStaleData struct {
	client     app.GithubClient
	store      KVStore
	ttl        time.Duration
	refreshTTL time.Duration
	l          logrus.FieldLogger

	projectUpdates chan projectsDBUpdateRequest
	statsUpdates   chan statsDBUpdateRequest

	// Chan for controlling scheduler - only used for unit testing.
	schedulerPendingOps chan int

	// Func for canceling internal worker loop and initializing db cleanup
	stop func()
}

// NewClientWithStaleData creates new ClientWithStaleData instance.
func NewClientWithStaleData(
	client app.GithubClient,
	store KVStore,
	ttl time.Duration,
	refreshTTL time.Duration,
	l logrus.FieldLogger,
) (*ClientWithStaleData, error) {
	c := ClientWithStaleData{
		client:         client,
		store:          store,
		ttl:            ttl,
		refreshTTL:     refreshTTL,
		l:              l,
		projectUpdates: make(chan projectsDBUpdateRequest, 1000),
		statsUpdates:   make(chan statsDBUpdateRequest, 1000),
	}

	return &c, nil
}

// RunScheduler runs internal scheduling goroutine.
// Doesn't block.
func (c *ClientWithStaleData) RunScheduler() {
	ctx, cancel := context.WithCancel(context.Background())
	c.stop = cancel

	go func() {
		pendingProjectUpdates := make(map[string]bool)
		pendingStatsUpdates := make(map[string]bool)

		doneProjectUpdates := make(chan string)
		doneStatsUpdates := make(chan string)

		for {
			// This is intended for blocking scheduler for unit testing.
			// In standard execution this is always nil.
			if c.schedulerPendingOps != nil {
				c.schedulerPendingOps <- len(pendingProjectUpdates) + len(pendingStatsUpdates)
			}

			select {
			// Projects
			case req := <-c.projectUpdates:
				if pendingProjectUpdates[req.language] {
					continue
				}
				pendingProjectUpdates[req.language] = true

				go func(req projectsDBUpdateRequest) {
					c.l.Infof("ClientWithStaleData: scheduled projects update for %s...", req.language)
					if err := c.updateProjects(req); err != nil {
						c.l.Errorf("ClientWithStaleData scheduler: updating projects data: %v", err)
					} else {
						c.l.Infof("ClientWithStaleData: scheduled projects update for %s done", req.language)
					}
					doneProjectUpdates <- req.language
				}(req)
			case key := <-doneProjectUpdates:
				delete(pendingProjectUpdates, key)

			// Stats
			case req := <-c.statsUpdates:
				key := fmt.Sprintf("%s/%s", req.owner, req.name)
				if pendingStatsUpdates[key] {
					continue
				}
				pendingStatsUpdates[key] = true

				go func(req statsDBUpdateRequest) {
					c.l.Infof("ClientWithStaleData: scheduled stats update for %s/%s...", req.owner, req.name)
					if err := c.updateStats(req); err != nil {
						c.l.Errorf("ClientWithStaleData scheduler: updating stats data: %v", err)
					} else {
						c.l.Infof("ClientWithStaleData: scheduled stats update for %s/%s done", req.owner, req.name)
					}
					doneStatsUpdates <- key
				}(req)
			case key := <-doneStatsUpdates:
				delete(pendingStatsUpdates, key)

			// Finish
			case <-ctx.Done():
				return
			}
		}
	}()
}

// ProjectsByLanguage returns projects by given programming language name.
//
// Returns data from db if available.
func (c *ClientWithStaleData) ProjectsByLanguage(ctx context.Context, language string, count int) ([]app.Project, error) {
	key := c.projectsDBKey(language)
	data, err := c.store.ReadKey(key)
	if err != nil {
		return nil, err
	}
	if data != nil {
		entry, err := c.unserializeProjects(data)
		if err != nil {
			return nil, fmt.Errorf("unserializing projects data: %w", err)
		}
		entryCreated := time.Unix(entry.Created, 0)
		if entry.Count >= count && entryCreated.Add(c.ttl).After(time.Now()) {
			if entryCreated.Add(c.refreshTTL).Before(time.Now()) {
				go func() {
					c.projectUpdates <- projectsDBUpdateRequest{
						language: language,
						count:    count,
					}
				}()
			}

			projects := entry.Data
			if len(projects) > count {
				projects = projects[:count]
			}
			return projects, nil
		}
	}

	select {
	case c.projectUpdates <- projectsDBUpdateRequest{
		language: language,
		count:    count,
	}:
		return nil, app.ScheduledForLaterError("scheduled")
	default:
		return nil, errors.New("stale data scheduler: no free slots left")
	}
}

// StatsByProject returns stats by given github project params.
//
// Returns data from db if available.
func (c *ClientWithStaleData) StatsByProject(ctx context.Context, name string, owner string) ([]app.ContributorStats, error) {
	key := c.statsDBKey(name, owner)
	data, err := c.store.ReadKey(key)
	if err != nil {
		return nil, err
	}
	if data != nil {
		entry, err := c.unserializeStats(data)
		if err != nil {
			return nil, fmt.Errorf("unserializing stats data: %w", err)
		}
		entryCreated := time.Unix(entry.Created, 0)
		if entryCreated.Add(c.ttl).After(time.Now()) {
			if entryCreated.Add(c.refreshTTL).Before(time.Now()) {
				c.statsUpdates <- statsDBUpdateRequest{
					name:  name,
					owner: owner,
				}
			}

			return entry.Data, nil
		}
	}

	select {
	case c.statsUpdates <- statsDBUpdateRequest{
		name:  name,
		owner: owner,
	}:
		return nil, app.ScheduledForLaterError("scheduled")
	default:
		return nil, errors.New("stale data scheduler: no free slots left")
	}
}

// Close cleanups scheduler and closes underlying database.
func (c *ClientWithStaleData) Close() {
	if c.stop != nil {
		c.stop()
		c.stop = nil
	}
}

func (c *ClientWithStaleData) updateProjects(req projectsDBUpdateRequest) error {
	projects, err := c.client.ProjectsByLanguage(context.Background(), req.language, req.count)
	if err != nil {
		return fmt.Errorf("calling client.ProjectsByLanguage: %w", err)
	}
	if err := c.saveProjects(req.language, req.count, projects); err != nil {
		return fmt.Errorf("saving projects: %w", err)
	}

	return nil
}

func (c *ClientWithStaleData) updateStats(req statsDBUpdateRequest) error {
	stats, err := c.client.StatsByProject(context.Background(), req.name, req.owner)
	if err != nil {
		return fmt.Errorf("calling client.StatsByProject: %w", err)
	}
	if err := c.saveStats(req.name, req.owner, stats); err != nil {
		return fmt.Errorf("saving stats: %w", err)
	}

	return nil
}

func (c *ClientWithStaleData) saveProjects(language string, count int, projects []app.Project) error {
	dbdata, err := c.serializeProjects(projectsDBEntry{
		Created: time.Now().Unix(),
		Count:   count,
		Data:    projects,
	})
	if err != nil {
		return fmt.Errorf("serializing data for save: %w", err)
	}

	return c.store.UpdateKey(c.projectsDBKey(language), dbdata)
}

func (c *ClientWithStaleData) saveStats(name string, owner string, stats []app.ContributorStats) error {
	dbdata, err := c.serializeStats(statsDBEntry{
		Created: time.Now().Unix(),
		Data:    stats,
	})
	if err != nil {
		return fmt.Errorf("serializing data for save: %w", err)
	}

	return c.store.UpdateKey(c.statsDBKey(name, owner), dbdata)
}

func (c *ClientWithStaleData) projectsDBKey(language string) []byte {
	return []byte("pr/" + language)
}

func (c *ClientWithStaleData) statsDBKey(name string, owner string) []byte {
	return []byte("st/" + owner + "/" + name)
}

func (c *ClientWithStaleData) serializeProjects(entry projectsDBEntry) ([]byte, error) {
	data, err := json.Marshal(entry)
	if err != nil {
		return nil, fmt.Errorf("marshalling json: %w", err)
	}

	return data, nil
}

func (c *ClientWithStaleData) unserializeProjects(data []byte) (*projectsDBEntry, error) {
	var entry projectsDBEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, fmt.Errorf("unmarshalling json: %w", err)
	}

	return &entry, nil
}

func (c *ClientWithStaleData) serializeStats(entry statsDBEntry) ([]byte, error) {
	data, err := json.Marshal(entry)
	if err != nil {
		return nil, fmt.Errorf("marshalling json: %w", err)
	}

	return data, nil
}

func (c *ClientWithStaleData) unserializeStats(data []byte) (*statsDBEntry, error) {
	var entry statsDBEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, fmt.Errorf("unmarshalling json: %w", err)
	}

	return &entry, nil
}

type projectsDBEntry struct {
	Created int64
	Count   int
	Data    []app.Project
}
type statsDBEntry struct {
	Created int64
	Data    []app.ContributorStats
}

type projectsDBUpdateRequest struct {
	language string
	count    int
}

type statsDBUpdateRequest struct {
	name  string
	owner string
}
