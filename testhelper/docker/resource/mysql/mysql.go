package mysql

import (
	"bytes"
	_ "encoding/json"
	"fmt"

	"github.com/ory/dockertest/v3"
	dc "github.com/ory/dockertest/v3/docker"

	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource"
	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource/internal"
	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource/registry"
)

const (
	defaultDB       = "sources"
	defaultUser     = "root"
	defaultPassword = "password"
)

type Resource struct {
	DBDsn    string
	Database string
	Password string
	User     string
	Host     string
	Port     string
}

func Setup(pool *dockertest.Pool, d resource.Cleaner, opts ...func(*Config)) (*Resource, error) {
	c := &Config{
		Tag: "8.2",
	}
	for _, opt := range opts {
		opt(c)
	}

	// pulls an image, creates a container based on it and runs it
	mysqlContainer, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: registry.ImagePath("mysql"),
		Tag:        c.Tag,
		Env: []string{
			"MYSQL_ROOT_PASSWORD=" + defaultPassword,
			"MYSQL_DATABASE=" + defaultDB,
		},
		ExposedPorts: []string{"3306/tcp"},
		PortBindings: internal.IPv4PortBindings([]string{"3306"}),
		Auth:         registry.AuthConfiguration(),
	}, func(hc *dc.HostConfig) {
		hc.ShmSize = c.ShmSize
	}, internal.DefaultHostConfig)
	if err != nil {
		return nil, fmt.Errorf("running container: %w", err)
	}

	d.Cleanup(func() {
		if err := pool.Purge(mysqlContainer); err != nil {
			d.Log("Could not purge resource:", err)
		}
	})

	dbDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?tls=false",
		defaultUser, defaultPassword,
		mysqlContainer.GetBoundIP("3306/tcp"),
		mysqlContainer.GetPort("3306/tcp"),
		defaultDB,
	)
	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	err = pool.Retry(func() (err error) {
		var w bytes.Buffer
		code, err := mysqlContainer.Exec([]string{
			"bash",
			"-c",
			"mysqladmin ping -h 127.0.0.1 --silent",
		}, dockertest.ExecOptions{StdOut: &w, StdErr: &w})
		if err != nil {
			return err
		}
		if code != 0 {
			return fmt.Errorf("mysql not ready:\n%s", w.String())
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("pinging container: %w", err)
	}
	return &Resource{
		DBDsn:    dbDSN,
		Database: defaultDB,
		User:     defaultUser,
		Password: defaultPassword,
		Host:     mysqlContainer.GetBoundIP("3306/tcp"),
		Port:     mysqlContainer.GetPort("3306/tcp"),
	}, nil
}
