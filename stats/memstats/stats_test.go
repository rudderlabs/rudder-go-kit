package memstats_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-go-kit/stats"
	"github.com/rudderlabs/rudder-go-kit/stats/memstats"
	"github.com/rudderlabs/rudder-go-kit/stats/testhelper/tracemodel"
)

func TestStats(t *testing.T) {
	now := time.Now()

	commonTags := stats.Tags{"tag1": "value1"}

	t.Run("test Count", func(t *testing.T) {
		name := "testCount"
		store, err := memstats.New(
			memstats.WithNow(func() time.Time {
				return now
			}),
		)
		require.NoError(t, err)

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
		store, err := memstats.New(
			memstats.WithNow(func() time.Time {
				return now
			}),
		)
		require.NoError(t, err)

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
		store, err := memstats.New(
			memstats.WithNow(func() time.Time {
				return now
			}),
		)
		require.NoError(t, err)

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
		store, err := memstats.New(
			memstats.WithNow(func() time.Time {
				return now
			}),
		)
		require.NoError(t, err)

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
		store, err := memstats.New(
			memstats.WithNow(func() time.Time {
				return now
			}),
		)
		require.NoError(t, err)

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
		store, err := memstats.New(
			memstats.WithNow(func() time.Time {
				return now
			}),
		)
		require.NoError(t, err)

		require.NoError(t, store.Start(context.Background(), stats.DefaultGoRoutineFactory))
		store.Stop()

		require.Equal(t, []memstats.Metric{}, store.GetAll())
	})

	t.Run("no tags", func(t *testing.T) {
		name := "no_tags"
		store, err := memstats.New(
			memstats.WithNow(func() time.Time {
				return now
			}),
		)
		require.NoError(t, err)

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

		store, err := memstats.New(
			memstats.WithNow(func() time.Time {
				return now
			}),
		)
		require.NoError(t, err)

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

	t.Run("with tracing", func(t *testing.T) {
		store, err := memstats.New(
			memstats.WithNow(func() time.Time {
				return now
			}),
			memstats.WithTracing(),
		)
		require.NoError(t, err)

		// we haven't done anything yet, so there should be no spans
		spans, err := store.Spans()
		require.NoError(t, err)
		require.Nil(t, spans)

		tracer := store.NewTracer("my-tracer")
		ctx, span1 := tracer.Start(context.Background(), "span1", stats.SpanKindInternal, stats.SpanWithTags(stats.Tags{
			"tag1": "value1",
			"tag2": "value2",
		}))

		_, span2 := tracer.Start(ctx, "span2", stats.SpanKindInternal, stats.SpanWithTags(stats.Tags{"tag3": "value3"}))
		time.Sleep(time.Millisecond)
		span2.End()
		time.Sleep(time.Millisecond)
		span1.End()

		_, unrelatedSpan := tracer.Start(
			context.Background(), "unrelatedSpan", stats.SpanKindInternal, stats.SpanWithTags(stats.Tags{
				"tag4": "value4",
			}),
		)
		time.Sleep(time.Millisecond)
		unrelatedSpan.End()

		spans, err = store.Spans()
		require.NoError(t, err)

		require.Len(t, spans, 3)
		require.Equal(t, "span2", spans[0].Name)
		require.Equal(t, "span1", spans[1].Name)
		require.Equal(t, "unrelatedSpan", spans[2].Name)
		require.True(t, spans[0].StartTime.IsZero())
		require.True(t, spans[1].StartTime.IsZero())
		require.True(t, spans[2].StartTime.IsZero())
		// checking hierarchy
		require.Equal(t, spans[1].SpanContext.SpanID, spans[0].Parent.SpanID)
		require.NotEmpty(t, spans[2].SpanContext.SpanID, spans[0].Parent.SpanID)
		require.NotEmpty(t, spans[2].SpanContext.SpanID, spans[1].Parent.SpanID)
		// checking attributes
		require.ElementsMatchf(t, []tracemodel.Attributes{{
			Key: "tag3",
			Value: tracemodel.Value{
				Type:  "STRING",
				Value: "value3",
			},
		}}, spans[0].Attributes, "span2 attributes: %+v", spans[0].Attributes)
		require.ElementsMatchf(t, []tracemodel.Attributes{{
			Key: "tag1",
			Value: tracemodel.Value{
				Type:  "STRING",
				Value: "value1",
			},
		}, {
			Key: "tag2",
			Value: tracemodel.Value{
				Type:  "STRING",
				Value: "value2",
			},
		}}, spans[1].Attributes, "span1 attributes: %+v", spans[1].Attributes)
		require.ElementsMatchf(t, []tracemodel.Attributes{{
			Key: "tag4",
			Value: tracemodel.Value{
				Type:  "STRING",
				Value: "value4",
			},
		}}, spans[2].Attributes, "unrelatedSpan attributes: %+v", spans[2].Attributes)
	})

	t.Run("with tracing timestamps", func(t *testing.T) {
		store, err := memstats.New(
			memstats.WithNow(func() time.Time {
				return now
			}),
			memstats.WithTracing(),
			memstats.WithTracingTimestamps(),
		)
		require.NoError(t, err)

		tracer := store.NewTracer("my-tracer")
		ctx, span1 := tracer.Start(context.Background(), "span1", stats.SpanKindInternal, stats.SpanWithTags(stats.Tags{
			"tag1": "value1",
			"tag2": "value2",
		}))

		_, span2 := tracer.Start(ctx, "span2", stats.SpanKindInternal, stats.SpanWithTags(stats.Tags{"tag3": "value3"}))
		span2.End()
		span1.End()

		spans, err := store.Spans()
		require.NoError(t, err)

		require.Len(t, spans, 2)
		// The data is extracted from stdout so the order is reversed
		require.Equal(t, "span2", spans[0].Name)
		require.Equal(t, "span1", spans[1].Name)
		require.False(t, spans[0].StartTime.IsZero())
		require.False(t, spans[1].StartTime.IsZero())
		// checking hierarchy
		require.Equal(t, spans[1].SpanContext.SpanID, spans[0].Parent.SpanID)
	})
}
