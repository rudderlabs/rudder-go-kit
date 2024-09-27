package scylla

import (
	"bytes"
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"

	"github.com/rudderlabs/rudder-go-kit/bytesize"
	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource"
	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource/internal"
	"github.com/rudderlabs/rudder-go-kit/testhelper/rand"
)

type Resource struct {
	URL  string
	Port string
}

func Setup(pool *dockertest.Pool, d resource.Cleaner, opts ...Option) (*Resource, error) {
	c := &config{tag: "6.1"}
	for _, opt := range opts {
		opt(c)
	}

	if c.network == nil {
		var err error
		c.network, err = pool.Client.CreateNetwork(docker.CreateNetworkOptions{
			Name:       "scylla_network_test_" + time.Now().Format("YY-MM-DD-") + rand.UniqueString(6),
			EnableIPv6: false,
		})
		if err != nil {
			return nil, err
		}
		d.Cleanup(func() {
			if err := pool.Client.RemoveNetwork(c.network.ID); err != nil {
				d.Logf("Error while removing Docker network: %v", err)
			}
		})
	}

	container, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository:   "scylladb/scylla",
		Tag:          c.tag,
		ExposedPorts: []string{"9042/tcp"},
		PortBindings: internal.IPv4PortBindings([]string{"9042"}),
		Cmd:          []string{"--smp 1"},
		NetworkID:    c.network.ID,
	}, internal.DefaultHostConfig, func(hc *docker.HostConfig) {
		hc.CPUCount = 1
		hc.Memory = 128 * bytesize.MB
	})
	if err != nil {
		return nil, err
	}

	d.Cleanup(func() {
		if err := pool.Purge(container); err != nil {
			d.Log("Could not purge resource:", err)
		}
	})

	url := fmt.Sprintf("%s:%s", container.GetBoundIP("9042/tcp"), container.GetPort("9042/tcp"))

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
		cluster.DisableInitialHostLookup = true
		session, err := cluster.CreateSession()
		if err != nil {
			return nil, err
		}
		defer session.Close()
		err = session.Query(fmt.Sprintf(
			"CREATE KEYSPACE IF NOT EXISTS %s WITH REPLICATION = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 };",
			c.keyspace,
		)).Exec()
		if err != nil {
			return nil, err
		}
	}

	return &Resource{
		URL:  url,
		Port: container.GetPort("9042/tcp"),
	}, nil
}
