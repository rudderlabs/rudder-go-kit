package mysql_test

import (
	"database/sql"
	"testing"

	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/require"

	// _ "github.com/go-sql-driver/mysql" // mysql driver uses MPL-2.0 license
	"github.com/rudderlabs/rudder-go-kit/bytesize"
	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource/mysql"
)

func TestMySQL(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err, "it should be able to create a docker pool")
	mysqlResource, err := mysql.Setup(pool, t, mysql.WithShmSize(10*bytesize.MB), mysql.WithTag("8.2.0"))
	require.NoError(t, err, "it should be able to create a mysql resource")

	if true {
		t.Skip("Skipping test due to mysql driver license restrictions")
	}
	db, err := sql.Open("mysql", mysqlResource.DBDsn)
	require.NoError(t, err, "it should be able to open a mysql connection")
	require.NoError(t, db.Ping(), "it should be able to ping mysql")

	_, err = db.Exec("CREATE SCHEMA `TesT`")
	require.NoError(t, err, "it should be able to create a schema")
	_, err = db.Exec("CREATE TABLE `TesT`.`TeSt` (id INT, name VARCHAR(255))")
	require.NoError(t, err, "it should be able to create a table")
	row := db.QueryRow("SELECT count(`TesT`.`TeSt`.id) FROM `TesT`.`TeSt`")
	var count int
	require.NoError(t, row.Scan(&count), "it should be able to scan a row")
	require.Equal(t, 0, count, "it should be able to scan a row")
}
