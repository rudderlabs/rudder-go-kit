package chiware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/rudderlabs/rudder-go-kit/stats"
)

const (
	defaultConcurrentRequestsUpdateInterval = 10 * time.Second
	minConcurrentRequestsUpdateInterval     = time.Second
)

type config struct {
	redactUnknownPaths               bool
	concurrentRequestsUpdateInterval time.Duration
}

type Option func(*config)

// RedactUnknownPaths sets the redactUnknownPaths flag.
// If set to true, the path will be redacted if the route is not found.
// If set to false, the path will be used as is.
func RedactUnknownPaths(redactUnknownPaths bool) Option {
	return func(c *config) {
		c.redactUnknownPaths = redactUnknownPaths
	}
}

func ConcurrentRequestsUpdateInterval(interval time.Duration) Option {
	return func(c *config) {
		c.concurrentRequestsUpdateInterval = interval
	}
}

func StatMiddleware(ctx context.Context, s stats.Stats, component string, options ...Option) func(http.Handler) http.Handler {
	conf := config{
		redactUnknownPaths:               true,
		concurrentRequestsUpdateInterval: defaultConcurrentRequestsUpdateInterval,
	}
	for _, option := range options {
		option(&conf)
	}
	if conf.concurrentRequestsUpdateInterval < minConcurrentRequestsUpdateInterval {
		conf.concurrentRequestsUpdateInterval = time.Second
	}

	var concurrentRequests atomic.Int64
	activeClientCount := s.NewStat(fmt.Sprintf("%s.concurrent_requests_count", component), stats.GaugeType)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(conf.concurrentRequestsUpdateInterval):
				activeClientCount.Gauge(concurrentRequests.Load())
			}
		}
	}()

	// getPath retrieves the path from the request.
	// The matched route's template is used if a match is found,
	// otherwise the request's URL path is used instead.
	getPath := func(r *http.Request) string {
		if path := chi.RouteContext(r.Context()).RoutePattern(); path != "" {
			return path
		}
		if conf.redactUnknownPaths {
			return "/redacted"
		}
		return r.URL.Path
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sw := newStatusCapturingWriter(w)
			start := time.Now()
			concurrentRequests.Add(1)
			defer concurrentRequests.Add(-1)

			next.ServeHTTP(sw, r)
			s.NewSampledTaggedStat(
				fmt.Sprintf("%s.response_time", component),
				stats.TimerType,
				map[string]string{
					"reqType": getPath(r),
					"method":  r.Method,
					"code":    strconv.Itoa(sw.status),
				}).Since(start)
		})
	}
}

// newStatusCapturingWriter returns a new, properly initialized statusCapturingWriter
func newStatusCapturingWriter(w http.ResponseWriter) *statusCapturingWriter {
	return &statusCapturingWriter{
		ResponseWriter: w,
		status:         http.StatusOK,
	}
}

// statusCapturingWriter is a response writer decorator that captures the status code.
type statusCapturingWriter struct {
	http.ResponseWriter
	status int
}

// WriteHeader override the http.ResponseWriter's `WriteHeader` method
func (w *statusCapturingWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}
