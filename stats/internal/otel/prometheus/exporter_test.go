package prometheus

import (
	"context"
	"strconv"
	"sync"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

// TestCollectWithConcurrentScrapes reproduces two Prometheus servers scraping the same endpoint while instruments are
// still being created, as when a service has just started. Every Collect records the metric names it has not seen yet,
// so overlapping scrapes write and read the same state: without a lock this is a data race under -race, and a
// "concurrent map read and map write" fatal error in production.
func TestCollectWithConcurrentScrapes(t *testing.T) {
	registry := prometheus.NewRegistry()
	exporter, err := New(WithRegisterer(registry), WithoutTargetInfo())
	require.NoError(t, err)
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(exporter))
	t.Cleanup(func() { require.NoError(t, provider.Shutdown(context.Background())) })
	meter := provider.Meter("test")

	const instruments, scrapers = 200, 4
	var wg sync.WaitGroup
	wg.Go(func() {
		for i := range instruments {
			counter, err := meter.Int64Counter("counter_" + strconv.Itoa(i))
			if err != nil {
				t.Error(err)
				return
			}
			counter.Add(context.Background(), 1)
		}
	})
	for range scrapers {
		wg.Go(func() {
			for range instruments {
				if _, err := registry.Gather(); err != nil {
					t.Error(err)
					return
				}
			}
		})
	}
	wg.Wait()

	families, err := registry.Gather()
	require.NoError(t, err)
	require.Len(t, families, instruments)
}
