package testhelper

import (
	"io"

	promclient "github.com/prometheus/client_model/go"
	promexpfmt "github.com/prometheus/common/expfmt"
	prommodel "github.com/prometheus/common/model"
)

// ParsePrometheusMetrics parses the given Prometheus metrics and returns a map of metric name to metric family.
func ParsePrometheusMetrics(rdr io.Reader) (map[string]*promclient.MetricFamily, error) {
	parser := promexpfmt.NewTextParser(prommodel.UTF8Validation)
	mf, err := parser.TextToMetricFamilies(rdr)
	if err != nil {
		return nil, err
	}
	return mf, nil
}
