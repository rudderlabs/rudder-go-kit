package gorillaware_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/rudderlabs/rudder-go-kit/gorillaware"
	"github.com/rudderlabs/rudder-go-kit/stats"
	"github.com/rudderlabs/rudder-go-kit/stats/mock_stats"
	"github.com/stretchr/testify/require"
)

func TestStatsMiddleware(t *testing.T) {
	component := "test"
	testCase := func(expectedStatusCode int, pathTemplate, requestPath, expectedReqType, expectedMethod string) func(t *testing.T) {
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
					"reqType": expectedReqType,
					"method":  expectedMethod,
					"code":    strconv.Itoa(expectedStatusCode),
				}).Return(measurement).Times(1)
			measurement.EXPECT().Since(gomock.Any()).Times(1)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			router := mux.NewRouter()
			router.Use(
				gorillaware.StatMiddleware(ctx, router, mockStats, component),
			)
			router.HandleFunc(pathTemplate, handler).Methods(expectedMethod)

			response := httptest.NewRecorder()
			request := httptest.NewRequest("GET", "http://example.com"+requestPath, http.NoBody)
			router.ServeHTTP(response, request)
			require.Equal(t, expectedStatusCode, response.Code)
		}
	}

	t.Run("template with param in path", testCase(http.StatusNotFound, "/v1/{param}", "/v1/abc", "/v1/{param}", "GET"))

	t.Run("template without param in path", testCase(http.StatusNotFound, "/v1/some-other/key", "/v1/some-other/key", "/v1/some-other/key", "GET"))
}
