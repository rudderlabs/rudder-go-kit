package memstats_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-go-kit/stats"
	"github.com/rudderlabs/rudder-go-kit/stats/memstats"
)

func TestStats(t *testing.T) {
	now := time.Now()

	store, err := memstats.New(
		memstats.WithNow(func() time.Time {
			return now
		}),
	)
	require.NoError(t, err)

	commonTags := stats.Tags{"tag1": "value1"}

	t.Run("test Count", func(t *testing.T) {
		name := "testCount"
		store := memstats.New(
			memstats.WithNow(func() time.Time {
				return now
			}),
		)
		m := store.NewTaggedStat(name, stats.CountType, commonTags)

		m.Increment()

		require.Equal(t, 1.0, store.Get(name, commonTags).LastValue())
		require.Equal(t, []float64{1.0}, store.Get(name, commonTags).Values())

		m.Count(2)

		require.Equal(t, 3.0, store.Get(name, commonTags).LastValue())
		require.Equal(t, []float64{1.0, 3.0}, store.Get(name, commonTags).Values())

		require.Equal(t, []memstats.Metric{{
			Name:  name,
			Tags:  commonTags,
			Value: 3.0,
		}}, store.GetAll())

		require.Equal(t, []memstats.Metric{{
			Name:  name,
			Tags:  commonTags,
			Value: 3.0,
		}}, store.GetByName(name))
	})

	t.Run("test Gauge", func(t *testing.T) {
		name := "testGauge"
		store := memstats.New(
			memstats.WithNow(func() time.Time {
				return now
			}),
		)
		m := store.NewTaggedStat(name, stats.GaugeType, commonTags)

		m.Gauge(1.0)

		require.Equal(t, 1.0, store.Get(name, commonTags).LastValue())
		require.Equal(t, []float64{1.0}, store.Get(name, commonTags).Values())

		m.Gauge(2.0)

		require.Equal(t, 2.0, store.Get(name, commonTags).LastValue())
		require.Equal(t, []float64{1.0, 2.0}, store.Get(name, commonTags).Values())

		require.Equal(t, []memstats.Metric{{
			Name:  name,
			Tags:  commonTags,
			Value: 2.0,
		}}, store.GetAll())

		require.Equal(t, []memstats.Metric{{
			Name:  name,
			Tags:  commonTags,
			Value: 2.0,
		}}, store.GetByName(name))
	})

	t.Run("test Histogram", func(t *testing.T) {
		name := "testHistogram"
		store := memstats.New(
			memstats.WithNow(func() time.Time {
				return now
			}),
		)

		m := store.NewTaggedStat(name, stats.HistogramType, commonTags)

		m.Observe(1.0)

		require.Equal(t, 1.0, store.Get(name, commonTags).LastValue())
		require.Equal(t, []float64{1.0}, store.Get(name, commonTags).Values())

		m.Observe(2.0)

		require.Equal(t, 2.0, store.Get(name, commonTags).LastValue())
		require.Equal(t, []float64{1.0, 2.0}, store.Get(name, commonTags).Values())

		require.Equal(t, []memstats.Metric{{
			Name:   name,
			Tags:   commonTags,
			Values: []float64{1.0, 2.0},
		}}, store.GetAll())

		require.Equal(t, []memstats.Metric{{
			Name:   name,
			Tags:   commonTags,
			Values: []float64{1.0, 2.0},
		}}, store.GetByName(name))
	})

	t.Run("test Timer", func(t *testing.T) {
		name := "testTimer"
		store := memstats.New(
			memstats.WithNow(func() time.Time {
				return now
			}),
		)

		m := store.NewTaggedStat(name, stats.TimerType, commonTags)

		m.SendTiming(time.Second)
		require.Equal(t, time.Second, store.Get(name, commonTags).LastDuration())
		require.Equal(t, []time.Duration{time.Second}, store.Get(name, commonTags).Durations())

		m.SendTiming(time.Minute)
		require.Equal(t, time.Minute, store.Get(name, commonTags).LastDuration())
		require.Equal(t,
			[]time.Duration{time.Second, time.Minute},
			store.Get(name, commonTags).Durations(),
		)

		func() {
			defer m.RecordDuration()()
			now = now.Add(time.Second)
		}()
		require.Equal(t, time.Second, store.Get(name, commonTags).LastDuration())
		require.Equal(t,
			[]time.Duration{time.Second, time.Minute, time.Second},
			store.Get(name, commonTags).Durations(),
		)

		m.Since(now.Add(-time.Minute))
		require.Equal(t, time.Minute, store.Get(name, commonTags).LastDuration())
		require.Equal(t,
			[]time.Duration{time.Second, time.Minute, time.Second, time.Minute},
			store.Get(name, commonTags).Durations(),
		)

		require.Equal(t, []memstats.Metric{{
			Name:      name,
			Tags:      commonTags,
			Durations: []time.Duration{time.Second, time.Minute, time.Second, time.Minute},
		}}, store.GetAll())

		require.Equal(t, []memstats.Metric{{
			Name:      name,
			Tags:      commonTags,
			Durations: []time.Duration{time.Second, time.Minute, time.Second, time.Minute},
		}}, store.GetByName(name))
	})

	t.Run("invalid operations", func(t *testing.T) {
		store := memstats.New(
			memstats.WithNow(func() time.Time {
				return now
			}),
		)

		require.PanicsWithValue(t, "operation Count not supported for measurement type:gauge", func() {
			store.NewTaggedStat("invalid_count", stats.GaugeType, commonTags).Count(1)
		})
		require.PanicsWithValue(t, "operation Increment not supported for measurement type:gauge", func() {
			store.NewTaggedStat("invalid_inc", stats.GaugeType, commonTags).Increment()
		})
		require.PanicsWithValue(t, "operation Gauge not supported for measurement type:count", func() {
			store.NewTaggedStat("invalid_gauge", stats.CountType, commonTags).Gauge(1)
		})
		require.PanicsWithValue(t, "operation SendTiming not supported for measurement type:histogram", func() {
			store.NewTaggedStat("invalid_send_timing", stats.HistogramType, commonTags).SendTiming(time.Second)
		})
		require.PanicsWithValue(t, "operation RecordDuration not supported for measurement type:histogram", func() {
			store.NewTaggedStat("invalid_record_duration", stats.HistogramType, commonTags).RecordDuration()
		})
		require.PanicsWithValue(t, "operation Since not supported for measurement type:histogram", func() {
			store.NewTaggedStat("invalid_since", stats.HistogramType, commonTags).Since(time.Now())
		})
		require.PanicsWithValue(t, "operation Observe not supported for measurement type:timer", func() {
			store.NewTaggedStat("invalid_observe", stats.TimerType, commonTags).Observe(1)
		})

		require.PanicsWithValue(t, "name cannot be empty", func() {
			store.GetByName("")
		})
	})

	t.Run("no op", func(t *testing.T) {
		store := memstats.New(
			memstats.WithNow(func() time.Time {
				return now
			}),
		)

		require.NoError(t, store.Start(context.Background(), stats.DefaultGoRoutineFactory))
		store.Stop()

		require.Equal(t, []memstats.Metric{}, store.GetAll())
	})

	t.Run("no tags", func(t *testing.T) {
		name := "no_tags"
		store := memstats.New(
			memstats.WithNow(func() time.Time {
				return now
			}),
		)

		m := store.NewStat(name, stats.CountType)

		m.Increment()

		require.Equal(t, 1.0, store.Get(name, nil).LastValue())

		require.Equal(t, []memstats.Metric{{
			Name:  name,
			Value: 1.0,
		}}, store.GetAll())

		require.Equal(t, []memstats.Metric{{
			Name:  name,
			Value: 1.0,
		}}, store.GetByName(name))
	})

	t.Run("get by name", func(t *testing.T) {
		name1 := "name_1"
		name2 := "name_2"

		store := memstats.New(
			memstats.WithNow(func() time.Time {
				return now
			}),
		)

		m1 := store.NewStat(name1, stats.CountType)
		m1.Increment()
		m2 := store.NewStat(name2, stats.TimerType)
		m2.SendTiming(time.Second)

		require.Equal(t, []memstats.Metric{{
			Name:  name1,
			Value: 1.0,
		}}, store.GetByName(name1))

		require.Equal(t, []memstats.Metric{{
			Name:      name2,
			Durations: []time.Duration{time.Second},
		}}, store.GetByName(name2))

		require.Equal(t, []memstats.Metric{{
			Name:  name1,
			Value: 1.0,
		}, {
			Name:      name2,
			Durations: []time.Duration{time.Second},
		}}, store.GetAll())
	})
}
