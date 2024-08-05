package scylla

import (
	"bytes"
	"fmt"
	"runtime"

	"github.com/gocql/gocql"
	"github.com/ory/dockertest/v3"

	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource"
)

type Resource struct {
	URL  string
	Port string
}

func Setup(pool *dockertest.Pool, d resource.Cleaner, opts ...Option) (*Resource, error) {
	c := &config{
		tag: "5.4.9",
	}
	for _, opt := range opts {
		opt(c)
	}
	goos := runtime.GOOS
	container, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository:   "scylladb/scylla",
		Tag:          c.tag,
		Env:          []string{},
		ExposedPorts: []string{"9042"},
	})
	if err != nil {
		return nil, err
	}

	d.Cleanup(func() {
		if err := pool.Purge(container); err != nil {
			d.Log("Could not purge resource:", err)
		}
	})

	var url string
	switch goos {
	case "darwin":
		url = fmt.Sprintf("localhost:%s", container.GetPort("9042/tcp"))
	default:
		url = fmt.Sprintf("172.17.0.2:%s", container.GetPort("9042/tcp"))
	}

	if err := pool.Retry(func() (err error) {
		var w bytes.Buffer
		code, err := container.Exec(
			[]string{
				"sh", "-c", "nodetool statusgossip | grep 'running' || exit 1",
			},
			dockertest.ExecOptions{StdOut: &w, StdErr: &w},
		)
		if err != nil {
			return err
		}
		if code != 0 {
			return fmt.Errorf("scylla healthcheck failed")
		}
		return nil
	}); err != nil {
		return nil, err
	}

	if err := pool.Retry(func() (err error) {
		var w bytes.Buffer
		code, err := container.Exec(
			[]string{
				"sh", "-c", "cqlsh || exit 1",
			},
			dockertest.ExecOptions{StdOut: &w, StdErr: &w},
		)
		if err != nil {
			return err
		}
		if code != 0 {
			return fmt.Errorf("scylla cql check failed")
		}
		return nil
	}); err != nil {
		return nil, err
	}

	if c.keyspace != "" {
		cluster := gocql.NewCluster(url)
		cluster.Consistency = gocql.Quorum
		session, err := cluster.CreateSession()
		if err != nil {
			return nil, err
		}
		defer session.Close()
		err = session.Query(fmt.Sprintf("CREATE KEYSPACE IF NOT EXISTS %s WITH REPLICATION = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 };", c.keyspace)).Exec()
		if err != nil {
			return nil, err
		}
	}

	return &Resource{
		URL:  url,
		Port: container.GetPort("9042/tcp"),
	}, nil
}
