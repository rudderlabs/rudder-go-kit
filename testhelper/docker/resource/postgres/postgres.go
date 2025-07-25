package postgres

import (
	"bytes"
	"database/sql"
	_ "encoding/json"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"

	"github.com/rudderlabs/rudder-go-kit/bytesize"
	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource"
	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource/internal"
)

const (
	postgresDefaultDB       = "jobsdb"
	postgresDefaultUser     = "rudder"
	postgresDefaultPassword = "password"
)

type Resource struct {
	DB       *sql.DB
	DBDsn    string
	Database string
	Password string
	User     string
	Host     string
	Port     string

	ContainerName string
	ContainerID   string
}

func Setup(pool *dockertest.Pool, d resource.Cleaner, opts ...func(*Config)) (*Resource, error) {
	c := &Config{
		Tag:     "17-alpine",
		ShmSize: 128 * bytesize.MB,
	}
	for _, opt := range opts {
		opt(c)
	}

	cmd := []string{"postgres"}
	for _, opt := range c.Options {
		cmd = append(cmd, "-c", opt)
	}
	// pulls an image, creates a container based on it and runs it
	postgresContainer, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        c.Tag,
		NetworkID:  c.NetworkID,
		Env: []string{
			"POSTGRES_PASSWORD=" + postgresDefaultPassword,
			"POSTGRES_DB=" + postgresDefaultDB,
			"POSTGRES_USER=" + postgresDefaultUser,
		},
		Cmd:          cmd,
		PortBindings: internal.IPv4PortBindings([]string{"5432"}, internal.WithBindIP(c.BindIP)),
	}, func(hc *docker.HostConfig) {
		hc.ShmSize = c.ShmSize
		hc.OOMKillDisable = c.OOMKillDisable
		hc.Memory = c.Memory
	}, internal.DefaultHostConfig)
	if err != nil {
		return nil, err
	}
	var db *sql.DB

	d.Cleanup(func() {
		if d.Failed() && c.PrintLogsOnError {
			if c, found := pool.ContainerByName(postgresContainer.Container.Name); found {
				d.Log(fmt.Sprintf("%q postgres container state: %+v", c.Container.Name, c.Container.State))
				b := bytes.NewBufferString("")
				if err := pool.Client.Logs(docker.LogsOptions{
					Container:    c.Container.ID,
					Stdout:       true,
					Stderr:       true,
					OutputStream: b,
					ErrorStream:  b,
				}); err != nil {
					_, _ = fmt.Fprintf(b, "could not get logs: %s", err)
				}
				d.Log(fmt.Sprintf("%q postgres container logs:\n%s", c.Container.Name, b.String()))
			}
		}
		if err := pool.Purge(postgresContainer); err != nil {
			d.Log("Could not purge resource:", err)
		}
		if db != nil {
			_ = db.Close()
		}
	})

	dbDSN := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		postgresDefaultUser, postgresDefaultPassword,
		postgresContainer.GetBoundIP("5432/tcp"),
		postgresContainer.GetPort("5432/tcp"),
		postgresDefaultDB,
	)

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	err = pool.Retry(func() (err error) {
		// 1. use pg_isready
		var w bytes.Buffer
		code, err := postgresContainer.Exec([]string{
			"bash",
			"-c",
			fmt.Sprintf("pg_isready -d %[1]s -U %[2]s", postgresDefaultDB, postgresDefaultUser),
		}, dockertest.ExecOptions{StdOut: &w, StdErr: &w})
		if err != nil {
			return err
		}
		if code != 0 {
			return fmt.Errorf("postgres not ready:\n%s", w.String())
		}

		// 2. create a sql.DB and verify connection
		if db, err = sql.Open("postgres", dbDSN); err != nil {
			return fmt.Errorf("opening database: %w", err)
		}
		defer func() {
			if err != nil {
				_ = db.Close()
			}
		}()
		if err = db.Ping(); err != nil {
			return fmt.Errorf("pinging database: %w", err)
		}
		var one int
		if err = db.QueryRow("SELECT 1").Scan(&one); err != nil {
			return fmt.Errorf("querying database: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("waiting for database to startup: %w", err)
	}
	return &Resource{
		DB:            db,
		DBDsn:         dbDSN,
		Database:      postgresDefaultDB,
		User:          postgresDefaultUser,
		Password:      postgresDefaultPassword,
		Host:          postgresContainer.GetBoundIP("5432/tcp"),
		Port:          postgresContainer.GetPort("5432/tcp"),
		ContainerName: postgresContainer.Container.Name,
		ContainerID:   postgresContainer.Container.ID,
	}, nil
}
