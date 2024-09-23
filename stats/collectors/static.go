package collectors

import "github.com/rudderlabs/rudder-go-kit/stats"

type staticStats struct {
	tags stats.Tags
	kv   map[string]uint64
}

// NewStaticMetric allows to capture a gauge metric that does not change during the lifetime of the application.
// Can be useful for capturing configuration values or application version.
func NewStaticMetric(key string, tags stats.Tags, value uint64) *staticStats {
	return &staticStats{
		tags: tags,
		kv: map[string]uint64{
			key: value,
		},
	}
}

func (s *staticStats) Collect(gaugeFunc func(key string, tag stats.Tags, val uint64)) {
	for k, v := range s.kv {
		gaugeFunc(k, s.tags, v)
	}
}

func (s *staticStats) Zero(gaugeFunc func(key string, tag stats.Tags, val uint64)) {
	for k := range s.kv {
		gaugeFunc(k, s.tags, 0)
	}
}
