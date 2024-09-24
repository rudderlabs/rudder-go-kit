package collectors_test

import (
	"testing"

	"github.com/rudderlabs/rudder-go-kit/stats"
	"github.com/rudderlabs/rudder-go-kit/stats/collectors"
	"github.com/rudderlabs/rudder-go-kit/stats/memstats"
	"github.com/stretchr/testify/require"
)

func TestStatic(t *testing.T) {
	testName := "test_sqlite"
	s := collectors.NewStaticMetric(testName, stats.Tags{
		"foo": "bar",
	}, 2)

	m, err := memstats.New()
	require.NoError(t, err)

	err = m.RegisterCollector(s)
	require.NoError(t, err)

	require.Equal(t, []memstats.Metric{
		{
			Name:  testName,
			Tags:  stats.Tags{"foo": "bar"},
			Value: 2,
		},
	}, m.GetAll())
}
