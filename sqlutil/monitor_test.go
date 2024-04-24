package sqlutil_test

import (
	"context"
	"testing"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-go-kit/config"
	"github.com/rudderlabs/rudder-go-kit/sqlutil"
	"github.com/rudderlabs/rudder-go-kit/stats"
	"github.com/rudderlabs/rudder-go-kit/stats/memstats"
	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource/postgres"
)

func TestMonitorDatabase(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)
	postgresContainer, err := postgres.Setup(pool, t)
	require.NoError(t, err)

	postgresContainer.DB.SetMaxOpenConns(10)
	postgresContainer.DB.SetMaxIdleConns(5)

	statsStore, err := memstats.New()
	require.NoError(t, err)

	identifier := "test"

	conf := config.New()
	conf.Set("Database.ReportInterval", "1s")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sqlutil.MonitorDatabase(ctx, conf, statsStore, postgresContainer.DB, identifier)

	require.Eventually(t, func() bool {
		return statsStore.Get("db_max_open_connections", stats.Tags{
			"identifier": identifier,
		}).LastValue() == 10
	},
		5*time.Second,
		100*time.Millisecond,
	)
	require.Eventually(t, func() bool {
		return statsStore.Get("db_open_connections", stats.Tags{
			"identifier": identifier,
		}).LastValue() == 1
	},
		5*time.Second,
		100*time.Millisecond,
	)
	require.Eventually(t, func() bool {
		return statsStore.Get("db_idle", stats.Tags{
			"identifier": identifier,
		}).LastValue() == 1
	},
		5*time.Second,
		100*time.Millisecond,
	)
}
