package postgres_test

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource/postgres"
)

func TestPostgres(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	for i := 1; i <= 6; i++ {
		t.Run(fmt.Sprintf("iteration %d", i), func(t *testing.T) {
			postgresContainer, err := postgres.Setup(pool, t)
			require.NoError(t, err)
			defer func() { _ = postgresContainer.DB.Close() }()

			db, err := sql.Open("postgres", postgresContainer.DBDsn)
			require.NoError(t, err)
			_, err = db.Exec("CREATE TABLE test (id int)")
			require.NoError(t, err)

			var count int
			err = db.QueryRow("SELECT count(*) FROM test").Scan(&count)
			require.NoError(t, err)
		})
	}

	t.Run("with test failure", func(t *testing.T) {
		cl := &testCleaner{T: t, failed: true}
		r, err := postgres.Setup(pool, cl)
		require.NoError(t, err)
		err = pool.Client.StopContainer(r.ContainerID, 10)
		require.NoError(t, err)
		cl.cleanup()
		require.Contains(t, cl.logs, "postgres container state: {Status:exited")
		require.Contains(t, cl.logs, "postgres container logs:")
	})
}

type testCleaner struct {
	*testing.T
	cleanup func()
	failed  bool
	logs    string
}

func (t *testCleaner) Cleanup(f func()) {
	t.cleanup = f
}

func (t *testCleaner) Failed() bool {
	return t.failed
}

func (t *testCleaner) Log(args ...any) {
	t.logs = t.logs + fmt.Sprint(args...)
}
