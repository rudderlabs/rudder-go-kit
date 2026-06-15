package stats

import (
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"

	"github.com/rudderlabs/rudder-go-kit/logger"
)

// TestCanonicalMeasurementIdentity exercises the sanitisation/exclusion rules and proves the returned
// attribute set preserves tag values losslessly (raw ':' and ',' survive — they are only sanitised for
// export, never for identity).
func TestCanonicalMeasurementIdentity(t *testing.T) {
	s := &otelStats{
		logger: logger.NOP,
		config: statsConfig{excludedTags: map[string]struct{}{
			"excludedRaw": {}, // excluded by raw key
			"a_b":         {}, // excluded by sanitized key (raw "a.b" sanitizes to this)
		}},
		resourceAttrs: map[string]struct{}{"namespace": {}},
	}

	for _, tc := range []struct {
		desc     string
		name     string
		tags     Tags
		wantName string
		wantTags map[string]string
	}{
		{
			desc: "blank name falls back to novalue", name: "   ", tags: nil,
			wantName: "novalue", wantTags: map[string]string{},
		},
		{
			desc: "plain tag survives", name: "lat", tags: Tags{"a": "1"},
			wantName: "lat", wantTags: map[string]string{"a": "1"},
		},
		{
			desc: "blank tag key is dropped", name: "lat", tags: Tags{"  ": "x", "a": "1"},
			wantName: "lat", wantTags: map[string]string{"a": "1"},
		},
		{
			desc: "tag excluded by raw key is dropped", name: "lat", tags: Tags{"excludedRaw": "x", "a": "1"},
			wantName: "lat", wantTags: map[string]string{"a": "1"},
		},
		{
			desc: "tag excluded by sanitized key is dropped", name: "lat", tags: Tags{"a.b": "x", "c": "1"},
			wantName: "lat", wantTags: map[string]string{"c": "1"},
		},
		{
			desc: "tag matching a resource attribute is dropped", name: "lat", tags: Tags{"namespace": "x", "a": "1"},
			wantName: "lat", wantTags: map[string]string{"a": "1"},
		},
		{
			desc: "key is sanitized but ':' is preserved", name: "lat", tags: Tags{"x.y": "1", "c:d": "2"},
			wantName: "lat", wantTags: map[string]string{"x_y": "1", "c:d": "2"},
		},
		{
			desc: "value is preserved raw, including ':' and ','", name: "lat", tags: Tags{"dest": "a:b,c-d"},
			wantName: "lat", wantTags: map[string]string{"dest": "a:b,c-d"},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			gotName, gotAttrs := s.canonicalMeasurementIdentity(tc.name, tc.tags)
			require.Equal(t, tc.wantName, gotName)
			require.Equal(t, tc.wantTags, attrsToMap(gotAttrs), "attribute set mirrors the surviving tags")
		})
	}
}

// TestCanonicalMeasurementIdentityDedup documents that two raw keys collapsing to the same sanitized key
// yield a single attribute (one of the two values wins — which one is map-iteration dependent, so only the
// key cardinality is asserted).
func TestCanonicalMeasurementIdentityDedup(t *testing.T) {
	s := &otelStats{logger: logger.NOP}
	_, gotAttrs := s.canonicalMeasurementIdentity("lat", Tags{"a.b": "1", "a_b": "2"})
	got := attrsToMap(gotAttrs)
	require.Len(t, got, 1, "both raw keys sanitize to a_b")
	require.Contains(t, []string{"1", "2"}, got["a_b"])
}

// TestMeasurementCacheKeyIsLossless is the core proof that the cache key (name + attribute identity)
// distinguishes series the OTel SDK treats as distinct — no lossy ':'→'-' or separator collapses.
// The key is attribute.Distinct (a 64-bit hash), so this asserts no collision for these concrete
// inputs; a genuine hash collision is not constructible in a test and is negligible in practice.
func TestMeasurementCacheKeyIsLossless(t *testing.T) {
	s := &otelStats{logger: logger.NOP, config: statsConfig{excludedTags: map[string]struct{}{"drop": {}}}}
	key := func(name string, tags Tags) measurementCacheKey {
		n, attrs := s.canonicalMeasurementIdentity(name, tags)
		return measurementCacheKey{n, attrs.Equivalent()}
	}

	t.Run("':' and '-' are distinct series", func(t *testing.T) {
		require.NotEqual(t, key("lat", Tags{"d": "a:b"}), key("lat", Tags{"d": "a-b"}))
	})
	t.Run("a comma in a value does not merge with an extra tag", func(t *testing.T) {
		// Old lossy key: both rendered as "a,1,b,2".
		require.NotEqual(t, key("lat", Tags{"a": "1,b,2"}), key("lat", Tags{"a": "1", "b": "2"}))
	})
	t.Run("nil and empty tags collapse to one key", func(t *testing.T) {
		require.Equal(t, key("lat", nil), key("lat", Tags{}))
	})
	t.Run("identity is independent of map construction order", func(t *testing.T) {
		require.Equal(t, key("lat", Tags{"a": "1", "b": "2"}), key("lat", Tags{"b": "2", "a": "1"}))
	})
	t.Run("an excluded tag does not change identity", func(t *testing.T) {
		require.Equal(t, key("lat", Tags{"a": "1"}), key("lat", Tags{"a": "1", "drop": "x"}))
	})
	t.Run("the same tags under different names are distinct", func(t *testing.T) {
		require.NotEqual(t, key("lat", Tags{"a": "1"}), key("size", Tags{"a": "1"}))
	})
	t.Run("a tag set that reduces to empty equals the no-tags key", func(t *testing.T) {
		// "  " is a blank key and is dropped, leaving an empty set — same identity as no tags at all.
		require.Equal(t, key("lat", nil), key("lat", Tags{"  ": "x"}))
	})
}

// TestMeasurementCacheWrapperIdentity proves, end to end through the public API, that distinct series get
// distinct cached wrappers while identical series share one — for every cached measurement type.
func TestMeasurementCacheWrapperIdentity(t *testing.T) {
	for _, statType := range []string{GaugeType, TimerType, HistogramType} {
		t.Run(statType, func(t *testing.T) {
			s := newOTelStats(t)
			same := func(a, b Measurement) { t.Helper(); require.Same(t, a, b) }
			distinct := func(a, b Measurement) { t.Helper(); require.NotSame(t, a, b) }

			// Happy path: identical name+tags reuse one wrapper (cache hit).
			same(
				s.NewTaggedStat("m", statType, Tags{"d": "x"}),
				s.NewTaggedStat("m", statType, Tags{"d": "x"}),
			)
			// NewStat, NewTaggedStat(nil) and NewTaggedStat(empty) are all the same (empty) series.
			same(s.NewStat("m", statType), s.NewTaggedStat("m", statType, nil))
			same(s.NewTaggedStat("m", statType, nil), s.NewTaggedStat("m", statType, Tags{}))
			// A dropped (blank) tag key does not change identity.
			same(
				s.NewTaggedStat("m", statType, Tags{"d": "x"}),
				s.NewTaggedStat("m", statType, Tags{"d": "x", "  ": "ignored"}),
			)

			// Unhappy path: ':' vs '-' must NOT collide (this is the bug being fixed).
			distinct(
				s.NewTaggedStat("m", statType, Tags{"d": "a:b"}),
				s.NewTaggedStat("m", statType, Tags{"d": "a-b"}),
			)
			// A comma in a value must not merge with a two-tag series.
			distinct(
				s.NewTaggedStat("m", statType, Tags{"a": "1,b,2"}),
				s.NewTaggedStat("m", statType, Tags{"a": "1", "b": "2"}),
			)
			// Different name, same tags → different wrapper.
			distinct(
				s.NewTaggedStat("m", statType, Tags{"d": "x"}),
				s.NewTaggedStat("n", statType, Tags{"d": "x"}),
			)
			// Different value, same key → different wrapper.
			distinct(
				s.NewTaggedStat("m", statType, Tags{"d": "x"}),
				s.NewTaggedStat("m", statType, Tags{"d": "y"}),
			)
		})
	}
}

// TestMeasurementCacheNoCrossContamination is the observable consequence of the wrapper identity: a write
// to one series of a lossy-colliding pair must never be visible on the other. One shape per type, across
// both collision shapes (':' vs '-' and an embedded separator).
func TestMeasurementCacheNoCrossContamination(t *testing.T) {
	const window = time.Minute

	t.Run("gauge keeps per-series values", func(t *testing.T) {
		s := newOTelStats(t)
		a := s.NewTaggedStat("g", GaugeType, Tags{"d": "a:b"})
		b := s.NewTaggedStat("g", GaugeType, Tags{"d": "a-b"})
		a.Gauge(1.0)
		b.Gauge(2.0)
		require.Equal(t, 1.0, a.(*otelGauge).getValue(), `dest="a:b" must keep its own value`)
		require.Equal(t, 2.0, b.(*otelGauge).getValue(), `dest="a-b" must keep its own value`)
	})

	t.Run("timer keeps per-series percentiles", func(t *testing.T) {
		s := newOTelStats(t)
		x := s.NewTaggedStat("dur", TimerType, Tags{"d": "a:b"})
		y := s.NewTaggedStat("dur", TimerType, Tags{"d": "a-b"})
		_, _ = x.Percentile(95, window) // enable both series
		_, ok := y.Percentile(95, window)
		require.False(t, ok, "y has no timings yet")
		for range 5 {
			x.SendTiming(2 * time.Second)
		}
		got, ok := y.Percentile(95, window)
		require.Falsef(t, ok, `dest="a-b" got no timings; reading dest="a:b"'s (p95=%v) is a cache collision`, got)
	})

	t.Run("histogram keeps per-series percentiles", func(t *testing.T) {
		s := newOTelStats(t)
		x := s.NewTaggedStat("h", HistogramType, Tags{"a": "1,b,2"})
		y := s.NewTaggedStat("h", HistogramType, Tags{"a": "1", "b": "2"})
		_, _ = x.Percentile(95, window) // enable both series
		_, ok := y.Percentile(95, window)
		require.False(t, ok, "y has no observations yet")
		for range 5 {
			x.Observe(7)
		}
		got, ok := y.Percentile(95, window)
		require.Falsef(t, ok, `two-tag series got no observations; reading the comma series' p95=%v is a collision`, got)
	})
}

// TestMeasurementCacheManyDistinctSeriesNoCollision mirrors the real usage where callers re-resolve a
// Measurement via NewTaggedStat on every observation instead of caching it, across many distinct series
// and many observations. Each series only ever observes its own unique value, so its min and max must
// equal that value; any cache-key collision would merge two series and corrupt at least one extreme.
func TestMeasurementCacheManyDistinctSeriesNoCollision(t *testing.T) {
	const (
		window    = time.Minute
		pairs     = 200 // 400 distinct series, adjacent ':' vs '-' twins (old-key collisions)
		perSeries = 100 // observations per series, each via a fresh NewTaggedStat lookup
	)
	s := newOTelStats(t)
	series := collidingSeriesPairs(pairs)
	value := func(i int) float64 { return float64(i) + 0.5 } // unique per series

	// The first Percentile call enables tracking for the series (inline lookup, no caching).
	for i := range series {
		_, _ = s.NewTaggedStat("obs", HistogramType, series[i]).Percentile(95, window)
	}
	// A lot of observations, every one through a fresh NewTaggedStat.
	for range perSeries {
		for i := range series {
			s.NewTaggedStat("obs", HistogramType, series[i]).Observe(value(i))
		}
	}
	// Each series must report exactly its own value at both extremes.
	for i := range series {
		m := s.NewTaggedStat("obs", HistogramType, series[i])
		lo, ok := m.Percentile(0, window)
		require.Truef(t, ok, "series %d %v has no data", i, series[i])
		hi, _ := m.Percentile(100, window)
		require.Equalf(t, value(i), lo, "series %d %v min contaminated", i, series[i])
		require.Equalf(t, value(i), hi, "series %d %v max contaminated", i, series[i])
	}
	// Exactly one cache entry per distinct series: fewer => a collision merged two; more => unstable keys.
	ostats := s.(*otelStats)
	ostats.histogramMeasurementsMu.Lock()
	cached := len(ostats.histogramMeasurements)
	ostats.histogramMeasurementsMu.Unlock()
	require.Equal(t, len(series), cached, "one cache entry per distinct series")
}

// TestMeasurementCacheKeyUniquenessHighCardinality probes the cache key directly over a large population
// of pairwise-distinct series, seeded with the exact shapes the old lossy key collapsed (':' vs '-', and
// a comma value vs a two-tag set). It catches structural collisions deterministically; it cannot catch a
// random 64-bit hash collision (not constructible, and ~k^2/2^65 is negligible for any real cardinality).
func TestMeasurementCacheKeyUniquenessHighCardinality(t *testing.T) {
	s := &otelStats{logger: logger.NOP}
	keyOf := func(tags Tags) measurementCacheKey {
		n, attrs := s.canonicalMeasurementIdentity("series", tags)
		return measurementCacheKey{n, attrs.Equivalent()}
	}

	const groups = 25000
	inputs := make([]Tags, 0, 4*groups)
	for i := range groups {
		inputs = append(
			inputs,
			Tags{"d": fmt.Sprintf("v%d:x", i)}, // ':' twin
			Tags{"d": fmt.Sprintf("v%d-x", i)}, // '-' twin (old key == ':' twin)
			Tags{"a": strconv.Itoa(i), "b": strconv.Itoa(i)}, // two tags
			Tags{"a": fmt.Sprintf("%d,b,%d", i, i)},          // comma value (old key == the two-tag form)
		)
	}

	seen := make(map[measurementCacheKey]Tags, len(inputs))
	for _, tags := range inputs {
		k := keyOf(tags)
		if prev, dup := seen[k]; dup {
			require.Failf(t, "cache-key collision", "distinct series share a key: %v vs %v", prev, tags)
		}
		seen[k] = tags
	}
	require.Len(t, seen, len(inputs), "every distinct series gets a unique key")
}

// TestMeasurementCacheConcurrentDistinctSeries first-touches many distinct series concurrently (each
// goroutine owns a disjoint slice), re-resolving NewTaggedStat per observation. Run with -race it also
// guards the cache's concurrent insert path. Each series observes only its own value, so none may bleed.
func TestMeasurementCacheConcurrentDistinctSeries(t *testing.T) {
	const (
		window       = time.Minute
		goroutines   = 8
		perGoroutine = 64
		observations = 50
	)
	s := newOTelStats(t)
	value := func(g, i int) float64 { return float64(g*1000 + i) }
	tagsFor := func(g, i int) Tags {
		return Tags{"g": strconv.Itoa(g), "i": fmt.Sprintf("n%d:x", i)} // ':' would collide under the old key
	}

	var wg sync.WaitGroup
	for g := range goroutines {
		wg.Go(func() {
			for i := range perGoroutine {
				_, _ = s.NewTaggedStat("c", HistogramType, tagsFor(g, i)).Percentile(95, window) // enable
			}
			for range observations {
				for i := range perGoroutine {
					s.NewTaggedStat("c", HistogramType, tagsFor(g, i)).Observe(value(g, i))
				}
			}
		})
	}
	wg.Wait()

	for g := range goroutines {
		for i := range perGoroutine {
			m := s.NewTaggedStat("c", HistogramType, tagsFor(g, i))
			lo, ok := m.Percentile(0, window)
			require.Truef(t, ok, "g=%d i=%d has no data", g, i)
			hi, _ := m.Percentile(100, window)
			require.Equalf(t, value(g, i), lo, "g=%d i=%d min contaminated", g, i)
			require.Equalf(t, value(g, i), hi, "g=%d i=%d max contaminated", g, i)
		}
	}
}

// BenchmarkMeasurementResolve quantifies the per-call cost of re-resolving a Measurement via
// NewTaggedStat on every observation (the common dev pattern) versus resolving it once and reusing it.
// The gap is the canonicalMeasurementIdentity work — tag sanitization + attribute.NewSet — paid on every
// call even though it is a cache hit. Run: go test -run '^$' -bench Resolve -benchmem.
func BenchmarkMeasurementResolve(b *testing.B) {
	s := newOTelStats(b)
	tags := Tags{"destinationId": "dest-123", "destType": "WEBHOOK", "status": "succeeded"}

	b.Run("re-resolve per observation", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			s.NewTaggedStat("events", HistogramType, tags).Observe(1)
		}
	})
	b.Run("cached measurement reused", func(b *testing.B) {
		m := s.NewTaggedStat("events", HistogramType, tags)
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			m.Observe(1)
		}
	})
}

// attrsToMap renders an attribute.Set as a plain map for assertions (all our attributes are strings).
func attrsToMap(set attribute.Set) map[string]string {
	m := make(map[string]string, set.Len())
	for _, kv := range set.ToSlice() {
		m[string(kv.Key)] = kv.Value.AsString()
	}
	return m
}

// collidingSeriesPairs builds 2*n pairwise-distinct tag sets as adjacent (':' , '-') twins.
// Under the old export-sanitized key ("a:b" and "a-b" both render "a-b") each twin pair collapsed to one entry;
// they must now stay distinct.
func collidingSeriesPairs(n int) []Tags {
	out := make([]Tags, 0, 2*n)
	for i := range n {
		out = append(
			out,
			Tags{"d": fmt.Sprintf("v%d:x", i)},
			Tags{"d": fmt.Sprintf("v%d-x", i)},
		)
	}
	return out
}
