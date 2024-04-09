package httputil_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-go-kit/httputil"
)

func TestGetRequestIP(t *testing.T) {
	testCases := []struct {
		name           string
		headerValue    string
		remoteAddr     string
		expectedResult string
	}{
		{
			name:           "X-Forwarded-For provided",
			headerValue:    "192.168.0.1, 192.168.0.2",
			remoteAddr:     "192.168.0.3:8080",
			expectedResult: "192.168.0.1",
		},
		{
			name:           "X-Forwarded-For empty, RemoteAddr provided",
			headerValue:    "",
			remoteAddr:     "192.168.0.4:8080",
			expectedResult: "192.168.0.4",
		},
		{
			name:           "X-Forwarded-For and RemoteAddr both empty",
			headerValue:    "",
			remoteAddr:     "",
			expectedResult: "",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			req := &http.Request{
				Header:     http.Header{"X-Forwarded-For": {testCase.headerValue}},
				RemoteAddr: testCase.remoteAddr,
			}
			require.Equal(t, testCase.expectedResult, httputil.GetRequestIP(req))
		})
	}
}
