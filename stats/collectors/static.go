package collectors

import (
	"fmt"

	"github.com/rudderlabs/rudder-go-kit/stats"
)

const (
	statsUniqName = "static_%s_%s"
)

type staticStats struct {
	tags  stats.Tags
	key   string
	value uint64
}

// NewStaticMetric allows to capture a gauge metric that does not change during the lifetime of the application.
// Can be useful for capturing configuration values or application version.
func NewStaticMetric(key string, tags stats.Tags, value uint64) *staticStats {
	return &staticStats{
		tags:  tags,
		key:   key,
		value: value,
	}
}

func (s *staticStats) Collect(gaugeFunc func(key string, tag stats.Tags, val uint64)) {
	gaugeFunc(s.key, s.tags, s.value)

}

func (s *staticStats) Zero(gaugeFunc func(key string, tag stats.Tags, val uint64)) {
	gaugeFunc(s.key, s.tags, 0)
}

func (s *staticStats) ID() string {
	return fmt.Sprintf(statsUniqName, s.key, s.tags.String())
}
