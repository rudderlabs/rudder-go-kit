package chiware_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-go-kit/chiware"
	"github.com/rudderlabs/rudder-go-kit/stats"
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
				chiware.StatMiddleware(ctx, router, mockStats, component, options...),
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
