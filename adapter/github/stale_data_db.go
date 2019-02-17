package github

import (
	"context"
	"encoding/json"
	"time"

	"github.com/etcd-io/bbolt"
	"github.com/m-zajac/goprojectdemo/app"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type ClientWithStaleData struct {
	client     app.GithubClient
	db         *bbolt.DB
	bucketName []byte
	ttl        time.Duration
	l          logrus.FieldLogger

	projectUpdates chan projectsDBUpdateRequest
	statsUpdates   chan projectsDBUpdateRequest

	// func for canceling internal worker loop and initializing db cleanup
	stop func()
}

// NewClientWithStaleData creates new ClientWithStaleData instance.
func NewClientWithStaleData(
	client app.GithubClient,
	dbPath string,
	bucketName string,
	ttl time.Duration,
	l logrus.FieldLogger,
) (*ClientWithStaleData, error) {
	db, err := bbolt.Open(dbPath, 0666, nil)
	if err != nil {
		return nil, errors.Wrap(err, "opening database")
	}

	if err = db.Update(func(tx *bbolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists([]byte(bucketName)); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "creating database bucket")
	}

	updatingCtx, updatingCtxCancel := context.WithCancel(context.Background())

	c := ClientWithStaleData{
		client:         client,
		db:             db,
		bucketName:     []byte(bucketName),
		ttl:            ttl,
		l:              l,
		projectUpdates: make(chan projectsDBUpdateRequest, 100),
		statsUpdates:   make(chan projectsDBUpdateRequest, 100),
		stop:           updatingCtxCancel,
	}

	go func() {
		for {
			select {
			case <-updatingCtx.Done():
				db.Close()
				return
			case req := <-c.projectUpdates:
				l.Info("background projects update!")

				var projects []app.Project
				if req.projects == nil {
					p, err := c.client.ProjectsByLanguage(context.Background(), req.language, req.count)
					if err != nil {
						l.Error(errors.Wrap(err, "ClientWithStaleData worker: calling client.ProjectsByLanguage"))
						continue
					}
					projects = p
				} else {
					projects = *req.projects
				}

				if err := c.writeProjects(req.language, req.count, projects); err != nil {
					l.Error(errors.Wrap(err, "ClientWithStaleData worker: writing projects"))
				}
				l.Info("background projects update done :)")
			}
		}
	}()

	return &c, nil
}

// ProjectsByLanguage returns projects by given programming language name.
//
// If data is available in db and ttl isn't exceeded, then data is returned and job for update is scheduled.
// Returned data is considered as "staled" in this case, but it would be eventually updated.
//
// If data is not available in db, client is called directly. Data returned from client is scheduled to save in db, then returned.
func (c *ClientWithStaleData) ProjectsByLanguage(ctx context.Context, language string, count int) ([]app.Project, error) {
	c.l.Info("HIT!")
	key := c.projectsDBKey(language)
	data, err := c.readKey(key)
	if err != nil {
		return nil, err
	}
	if data != nil {
		entry, err := c.unserializeProjects(data)
		if err != nil {
			return nil, errors.Wrap(err, "unserializing projects data")
		}
		entryCreated := time.Unix(entry.Created, 0)
		if entry.Count >= count && entryCreated.Add(c.ttl).After(time.Now()) {
			go func() {
				c.projectUpdates <- projectsDBUpdateRequest{
					language: language,
					count:    count,
				}
			}()

			projects := entry.Data
			if len(projects) > count {
				projects = projects[:count]
			}
			c.l.Info("returning from db")
			return projects, nil
		}
	}

	projects, err := c.client.ProjectsByLanguage(ctx, language, count)
	if err != nil {
		return projects, err
	}

	go func() {
		c.projectUpdates <- projectsDBUpdateRequest{
			language: language,
			count:    count,
			projects: &projects,
		}
	}()

	return projects, nil
}

// StatsByProject returns stats by given github project params.
func (c *ClientWithStaleData) StatsByProject(ctx context.Context, name string, owner string) ([]app.ContributorStats, error) {
	// TODO
	stats, err := c.client.StatsByProject(ctx, name, owner)
	if err != nil {
		return stats, err
	}

	return stats, nil
}

func (c *ClientWithStaleData) Close() {
	c.stop()
}

func (c *ClientWithStaleData) writeProjects(language string, count int, projects []app.Project) error {
	dbdata, err := c.serializeProjects(projectsDBEntry{
		Created: time.Now().Unix(),
		Count:   count,
		Data:    projects,
	})
	if err != nil {
		return errors.Wrap(err, "serializing data for save")
	}

	return c.writeKey(c.projectsDBKey(language), dbdata)
}

func (c *ClientWithStaleData) projectsDBKey(language string) []byte {
	return []byte(language)
}

func (c *ClientWithStaleData) statsDBKey(name string, owner string) []byte {
	return []byte(name + "/" + owner)
}

func (c *ClientWithStaleData) readKey(key []byte) ([]byte, error) {
	var data []byte
	if err := c.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(c.bucketName)
		data = b.Get(key)
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "reading from db")
	}

	return data, nil
}

func (c *ClientWithStaleData) writeKey(key []byte, data []byte) error {
	if err := c.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(c.bucketName)
		return b.Put(key, data)
	}); err != nil {
		return errors.Wrap(err, "writing to db")
	}

	return nil
}

func (c *ClientWithStaleData) serializeProjects(entry projectsDBEntry) ([]byte, error) {
	data, err := json.Marshal(entry)
	if err != nil {
		return nil, errors.Wrap(err, "marshalling json")
	}

	return data, nil
}

func (c *ClientWithStaleData) unserializeProjects(data []byte) (*projectsDBEntry, error) {
	var entry projectsDBEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, errors.Wrap(err, "unmarshalling json")
	}

	return &entry, nil
}

type projectsDBEntry struct {
	Created int64
	Count   int
	Data    []app.Project
}

type projectsDBUpdateRequest struct {
	language string
	count    int
	projects *[]app.Project
}
