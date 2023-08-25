package sqlutil_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-go-kit/sqlutil"
	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource"
)

func TestPrintRowsToTable(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)
	postgres, err := resource.SetupPostgres(pool, t)
	require.NoError(t, err)

	_, err = postgres.DB.Exec(`CREATE TABLE users (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		age INT NOT NULL,
		info JSONB
		)`)
	require.NoError(t, err)
	_, err = postgres.DB.Exec(`INSERT INTO users (name, age, info) VALUES ('John Doe', 20, '{"email": "jdoe@example.com"}')`)
	require.NoError(t, err)
	_, err = postgres.DB.Exec(`INSERT INTO users (name, age, info) VALUES ('Eva Chung', 20, '{"email": "echung@example.com"}')`)
	require.NoError(t, err)

	var out bytes.Buffer
	rows, err := postgres.DB.Query(`SELECT * FROM users`)
	require.NoError(t, err)
	defer func() { _ = rows.Close() }()
	err = sqlutil.PrintRowsToTable(rows, &out)
	require.NoError(t, err)
	fmt.Println(out.String())
	require.Equal(t,
		` |  id|      name| age|                            info|
 | ---|       ---| ---|                             ---|
 |   1|  John Doe|  20|   {"email": "jdoe@example.com"}|
 |   2| Eva Chung|  20| {"email": "echung@example.com"}|
`, out.String())
}
