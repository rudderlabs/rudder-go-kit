package jaeger

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ory/dockertest/v3"

	"github.com/rudderlabs/rudder-go-kit/httputil"
	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource"
	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource/internal"
	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource/registry"
)

const (
	otlpHTTPPort = "4318"
	queryPort    = "16686"
)

type Resource struct {
	// OTLPEndpoint is the OTLP HTTP endpoint (host:port format, no scheme)
	OTLPEndpoint string
	// QueryURL is the Jaeger query API URL
	QueryURL string

	pool     *dockertest.Pool
	resource *dockertest.Resource
}

func Setup(pool *dockertest.Pool, d resource.Cleaner) (*Resource, error) {
	jaeger, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository:   registry.ImagePath("jaegertracing/all-in-one"),
		Tag:          "1.68.0",
		ExposedPorts: []string{otlpHTTPPort + "/tcp", queryPort + "/tcp"},
		PortBindings: internal.IPv4PortBindings([]string{otlpHTTPPort, queryPort}),
		Auth:         registry.AuthConfiguration(),
	}, internal.DefaultHostConfig)
	if err != nil {
		return nil, fmt.Errorf("starting jaeger: %w", err)
	}

	res := &Resource{
		pool:         pool,
		resource:     jaeger,
		OTLPEndpoint: fmt.Sprintf("%s:%s", jaeger.GetBoundIP(otlpHTTPPort+"/tcp"), jaeger.GetPort(otlpHTTPPort+"/tcp")),
		QueryURL:     fmt.Sprintf("http://%s:%s", jaeger.GetBoundIP(queryPort+"/tcp"), jaeger.GetPort(queryPort+"/tcp")),
	}

	if jaeger.GetBoundIP(otlpHTTPPort+"/tcp") == "" {
		return nil, fmt.Errorf("getting jaeger bound ip")
	}

	d.Cleanup(func() {
		if err := pool.Purge(jaeger); err != nil {
			d.Log("Could not purge jaeger resource:", err)
		}
	})

	healthReq, err := http.NewRequest(http.MethodGet, res.QueryURL+"/", nil)
	if err != nil {
		return nil, fmt.Errorf("creating jaeger health request: %w", err)
	}

	err = pool.Retry(func() error {
		resp, err := http.DefaultClient.Do(healthReq)
		if err != nil {
			return fmt.Errorf("getting jaeger health: %w", err)
		}

		defer func() { httputil.CloseResponse(resp) }()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("jaeger health returned status code %d", resp.StatusCode)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("waiting for jaeger to be ready: %w", err)
	}

	return res, nil
}

// Trace represents a Jaeger trace from the query API
type Trace struct {
	TraceID string `json:"traceID"`
	Spans   []Span `json:"spans"`
}

// Span represents a span in a Jaeger trace
type Span struct {
	TraceID       string    `json:"traceID"`
	SpanID        string    `json:"spanID"`
	OperationName string    `json:"operationName"`
	References    []SpanRef `json:"references"`
	Tags          []Tag     `json:"tags"`
	Logs          []Log     `json:"logs"`
}

// SpanRef represents a reference to another span
type SpanRef struct {
	RefType string `json:"refType"`
	TraceID string `json:"traceID"`
	SpanID  string `json:"spanID"`
}

// Tag represents a span tag
type Tag struct {
	Key   string `json:"key"`
	Type  string `json:"type"`
	Value any    `json:"value"`
}

// Log represents a span log/event
type Log struct {
	Timestamp int64 `json:"timestamp"`
	Fields    []Tag `json:"fields"`
}

// QueryResponse represents the Jaeger query API response
type QueryResponse struct {
	Data   []Trace `json:"data"`
	Errors []any   `json:"errors"`
}

// GetTraces fetches traces for a given service from Jaeger
func (r *Resource) GetTraces(serviceName string) ([]Trace, error) {
	url := fmt.Sprintf("%s/api/traces?service=%s", r.QueryURL, serviceName)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching traces: %w", err)
	}
	defer func() { httputil.CloseResponse(resp) }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("jaeger returned status code %d", resp.StatusCode)
	}

	var queryResp QueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&queryResp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return queryResp.Data, nil
}
