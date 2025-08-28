package etcd

import (
	"context"
	"fmt"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	etcd "go.etcd.io/etcd/client/v3"

	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource"
	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource/internal"
)

type Resource struct {
	Client *etcd.Client
	Hosts  []string
	// HostsInNetwork is the list of ETCD hosts accessible from the provided Docker network (if any).
	HostsInNetwork []string
}

type config struct {
	network *docker.Network
}

type Option func(*config)

func WithNetwork(network *docker.Network) Option {
	return func(c *config) {
		c.network = network
	}
}

func Setup(pool *dockertest.Pool, cln resource.Cleaner, opts ...Option) (*Resource, error) {
	var c config
	for _, opt := range opts {
		opt(&c)
	}

	var networkID string
	if c.network != nil {
		networkID = c.network.ID
	}
	container, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "bitnamilegacy/etcd",
		Tag:        "3.5",
		NetworkID:  networkID,
		Env: []string{
			"ALLOW_NONE_AUTHENTICATION=yes",
		},
		PortBindings: internal.IPv4PortBindings([]string{"2379"}),
	}, internal.DefaultHostConfig)
	cln.Cleanup(func() {
		if err := pool.Purge(container); err != nil {
			cln.Log(fmt.Errorf("could not purge ETCD resource: %v", err))
		}
	})
	if err != nil {
		return nil, fmt.Errorf("could not create container: %v", err)
	}

	var (
		etcdClient *etcd.Client
		etcdHosts  []string
	)
	etcdHosts = []string{"http://" + container.GetBoundIP("2379/tcp") + ":" + container.GetPort("2379/tcp")}

	etcdClient, err = etcd.New(etcd.Config{
		Endpoints:   etcdHosts,
		DialTimeout: time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("setting up etcd client: %w", err)
	}

	err = pool.Retry(func() (err error) {
		_, err = etcdClient.MemberList(context.Background())

		return err
	})
	if err != nil {
		return nil, fmt.Errorf("could not connect to dockerized ETCD: %v", err)
	}

	var hostsInNetwork []string
	if c.network != nil {
		hostsInNetwork = []string{
			"http://" + container.GetIPInNetwork(&dockertest.Network{Network: c.network}) + ":2379",
		}
	}

	return &Resource{
		Client:         etcdClient,
		Hosts:          etcdHosts,
		HostsInNetwork: hostsInNetwork,
	}, nil
}
