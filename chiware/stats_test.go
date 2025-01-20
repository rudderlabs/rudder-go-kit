package chiware_test

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/rudderlabs/rudder-go-kit/chiware"
	"github.com/rudderlabs/rudder-go-kit/stats"
	"github.com/rudderlabs/rudder-go-kit/stats/memstats"
	"github.com/rudderlabs/rudder-go-kit/stats/mock_stats"
)

func TestStatsMiddleware(t *testing.T) {
	component := "test"
	testCase := func(expectedStatusCode int, pathTemplate, requestPath, expectedMethod string, options ...chiware.Option) func(t *testing.T) {
		return func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockStats := mock_stats.NewMockStats(ctrl)
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(expectedStatusCode)
			})

			measurement := mock_stats.NewMockMeasurement(ctrl)
			mockStats.EXPECT().NewStat(fmt.Sprintf("%s.concurrent_requests_count", component), stats.GaugeType).Return(measurement).Times(1)
			mockStats.EXPECT().NewSampledTaggedStat(fmt.Sprintf("%s.response_time", component), stats.TimerType,
				map[string]string{
					"reqType": pathTemplate,
					"method":  expectedMethod,
					"code":    strconv.Itoa(expectedStatusCode),
				}).Return(measurement).Times(1)
			measurement.EXPECT().Since(gomock.Any()).Times(1)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			router := chi.NewRouter()
			router.Use(
				chiware.StatMiddleware(ctx, mockStats, component, options...),
			)
			router.MethodFunc(expectedMethod, pathTemplate, handler)

			response := httptest.NewRecorder()
			request := httptest.NewRequest("GET", "http://example.com"+requestPath, http.NoBody)
			router.ServeHTTP(response, request)
			require.Equal(t, expectedStatusCode, response.Code)
		}
	}

	t.Run("template with param in path", testCase(http.StatusNotFound, "/v1/{param}", "/v1/abc", "GET"))
	t.Run("template without param in path", testCase(http.StatusNotFound, "/v1/some-other/key", "/v1/some-other/key", "GET"))
	t.Run("template with unknown path ", testCase(http.StatusNotFound, "/a/b/c", "/a/b/c", "GET", chiware.RedactUnknownPaths(false)))
	t.Run("template with unknown path ", testCase(http.StatusNotFound, "/redacted", "/a/b/c", "GET", chiware.RedactUnknownPaths(true)))
	t.Run("template with unknown path ", testCase(http.StatusNotFound, "/redacted", "/a/b/c", "GET"))
}

func TestVerifyConcurrency(t *testing.T) {
	ctx := context.Background()
	ms, err := memstats.New()
	require.NoError(t, err)

	router := chi.NewRouter()
	router.Use(
		chiware.StatMiddleware(ctx, ms, "test", chiware.ConcurrentRequestsUpdateInterval(1*time.Second)),
	)
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})

	srv := httptest.NewServer(router)
	defer srv.Close()

	var (
		wg             sync.WaitGroup
		maxConcurrency = 1000
		guard          = make(chan struct{}, maxConcurrency)
		timeout        = time.After(10 * time.Second)
	)

	go func() {
		ticker := time.NewTicker(time.Second)
		for {
			<-ticker.C

			metric := ms.Get("test.concurrent_requests_count", stats.Tags{})
			t.Logf("Concurrency: %.0f", metric.LastValue())
		}
	}()

loop:
	for {
		select {
		case <-timeout:
			break loop
		case guard <- struct{}{}:
			wg.Add(1)
			go func() {
				defer func() {
					wg.Done()
					<-guard
				}()
				resp, err := http.Get(srv.URL)
				require.NoError(t, err)
				require.Equal(t, http.StatusOK, resp.StatusCode)
			}()
		}
	}

	wg.Wait()
}
