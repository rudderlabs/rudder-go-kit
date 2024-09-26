package collectors_test

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-go-kit/stats"
	"github.com/rudderlabs/rudder-go-kit/stats/collectors"
	"github.com/rudderlabs/rudder-go-kit/stats/memstats"
)

func TestSQLDatabase(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(5)

	m, err := memstats.New()
	require.NoError(t, err)

	testName := "test_sqlite"
	s := collectors.NewDatabaseSQLStats(testName, db)

	err = m.RegisterCollector(s)
	require.NoError(t, err)

	require.Equal(t, []memstats.Metric{
		{
			Name:  "sql_db_idle_connections",
			Tags:  stats.Tags{"name": testName},
			Value: 1,
		},
		{
			Name:  "sql_db_in_use_connections",
			Tags:  stats.Tags{"name": testName},
			Value: 0,
		},
		{
			Name:  "sql_db_max_idle_closed_total",
			Tags:  stats.Tags{"name": testName},
			Value: 0,
		},
		{
			Name:  "sql_db_max_idle_time_closed_total",
			Tags:  stats.Tags{"name": testName},
			Value: 0,
		},
		{
			Name:  "sql_db_max_lifetime_closed_total",
			Tags:  stats.Tags{"name": testName},
			Value: 0,
		},
		{
			Name:  "sql_db_max_open_connections",
			Tags:  stats.Tags{"name": testName},
			Value: 5,
		},
		{
			Name:  "sql_db_open_connections",
			Tags:  stats.Tags{"name": testName},
			Value: 1,
		},
		{
			Name:  "sql_db_wait_count_total",
			Tags:  stats.Tags{"name": testName},
			Value: 0,
		},
		{
			Name:  "sql_db_wait_duration_seconds_total",
			Tags:  stats.Tags{"name": testName},
			Value: 0,
		},
	}, m.GetAll())
}
