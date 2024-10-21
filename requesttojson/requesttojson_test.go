package requesttojson

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRequestToJSON_SimpleGET(t *testing.T) {
	// Create a simple GET request
	req, err := http.NewRequest("GET", "http://example.com/path?query=1", nil)
	require.NoError(t, err)

	// Call the function
	jsonData, reqJSON, err := RequestToJSON(req, "")
	require.NoError(t, err)

	// Check the struct fields
	require.Equal(t, "GET", reqJSON.Method)
	require.Equal(t, "/path?query=1", reqJSON.URL)
	require.Equal(t, "HTTP/1.1", reqJSON.Proto)
	require.Empty(t, reqJSON.Body)
	require.Equal(t, map[string][]string{"query": {"1"}}, reqJSON.Query)

	// Check the JSON output
	require.Contains(t, string(jsonData), `"method":"GET"`)
	require.Contains(t, string(jsonData), `"url":"/path?query=1"`)
	require.Contains(t, string(jsonData), `"proto":"HTTP/1.1"`)
	require.Contains(t, string(jsonData), `"query_parameters":{"query":["1"]}`)
}

func TestRequestToJSON_WithHeaders(t *testing.T) {
	// Create a POST request with headers
	req, err := http.NewRequest("POST", "http://example.com/path", nil)
	require.NoError(t, err)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Custom-Header", "CustomValue")

	// Call the function
	jsonData, reqJSON, err := RequestToJSON(req, "")
	require.NoError(t, err)

	// Check the headers in struct
	require.Equal(t, map[string][]string{
		"Content-Type":    {"application/json"},
		"X-Custom-Header": {"CustomValue"},
	}, reqJSON.Headers)

	// Check the JSON output
	require.Contains(t, string(jsonData), `"Content-Type":["application/json"]`)
	require.Contains(t, string(jsonData), `"X-Custom-Header":["CustomValue"]`)
}

func TestRequestToJSON_WithBody(t *testing.T) {
	// Create a POST request with a body
	body := `{"key": "value"}`
	req, err := http.NewRequest("POST", "http://example.com/path", strings.NewReader(body))
	require.NoError(t, err)

	// Call the function
	jsonData, reqJSON, err := RequestToJSON(req, "")
	require.NoError(t, err)

	// Check the body
	require.Equal(t, body, reqJSON.Body)

	// Check the JSON output
	require.Contains(t, string(jsonData), `"body":"{\"key\": \"value\"}"`)
}

func TestRequestToJSON_WithQueryParams(t *testing.T) {
	// Create a GET request with query parameters
	req, err := http.NewRequest("GET", "http://example.com/search?q=golang&sort=asc", nil)
	require.NoError(t, err)

	// Call the function
	jsonData, reqJSON, err := RequestToJSON(req, "")
	require.NoError(t, err)

	// Check the query parameters
	expectedQuery := map[string][]string{
		"q":    {"golang"},
		"sort": {"asc"},
	}
	require.Equal(t, expectedQuery, reqJSON.Query)

	// Check the JSON output
	require.Contains(t, string(jsonData), `"query_parameters":{"q":["golang"],"sort":["asc"]}`)
}

func TestRequestToJSON_EmptyBodyReplaced(t *testing.T) {
	// Create a POST request with an empty body
	req, err := http.NewRequest("POST", "http://example.com/path", nil)
	require.NoError(t, err)

	// Call the function
	jsonData, reqJSON, err := RequestToJSON(req, "{}")
	require.NoError(t, err)

	// Check the body is empty
	require.Equal(t, reqJSON.Body, "{}")

	// Check the JSON output
	require.Contains(t, string(jsonData), `"body":"{}"`)
}

func TestRequestToJSON_EmptyBody(t *testing.T) {
	// Create a POST request with an empty body
	req, err := http.NewRequest("POST", "http://example.com/path", nil)
	require.NoError(t, err)

	// Call the function
	jsonData, reqJSON, err := RequestToJSON(req, "")
	require.NoError(t, err)

	// Check the body is empty
	require.Empty(t, reqJSON.Body)

	// Check the JSON output
	require.Contains(t, string(jsonData), `"body":""`)
}

func TestRequestToJSON_VariousHTTPMethods(t *testing.T) {
	methods := []string{"PUT", "DELETE", "PATCH", "OPTIONS", "HEAD"}

	for _, method := range methods {
		// Create a request with different HTTP methods
		req, err := http.NewRequest(method, "http://example.com/path", nil)
		require.NoError(t, err)

		// Call the function
		jsonData, reqJSON, err := RequestToJSON(req, "")
		require.NoError(t, err)

		// Check the method is correctly set
		require.Equal(t, method, reqJSON.Method)
		require.Contains(t, string(jsonData), `"method":"`+method+`"`)
	}
}

func TestRequestToJSON_URLEncodedBody(t *testing.T) {
	// Create a POST request with URL-encoded data
	body := "name=John+Doe&age=30"
	req, err := http.NewRequest("POST", "http://example.com/form", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// Call the function
	jsonData, reqJSON, err := RequestToJSON(req, "")
	require.NoError(t, err)

	// Check that the body is treated as a raw string
	require.Equal(t, body, reqJSON.Body)
	require.Contains(t, string(jsonData), `"body":"name=John+Doe\u0026age=30"`)
	require.Contains(t, string(jsonData), `"Content-Type":["application/x-www-form-urlencoded"]`)
}

func TestRequestToJSON_MultipleHeadersWithSameKey(t *testing.T) {
	// Create a GET request with repeated headers
	req, err := http.NewRequest("GET", "http://example.com/path", nil)
	require.NoError(t, err)
	req.Header.Add("X-Forwarded-For", "192.168.1.1")
	req.Header.Add("X-Forwarded-For", "10.0.0.1")

	// Call the function
	jsonData, reqJSON, err := RequestToJSON(req, "")
	require.NoError(t, err)

	// Check the headers in the struct
	expectedHeaders := map[string][]string{
		"X-Forwarded-For": {"192.168.1.1", "10.0.0.1"},
	}
	require.Equal(t, expectedHeaders, reqJSON.Headers)

	// Check the JSON output
	require.Contains(t, string(jsonData), `"X-Forwarded-For":["192.168.1.1","10.0.0.1"]`)
}

func TestRequestToJSON_InvalidJSONBody(t *testing.T) {
	// Create a POST request with invalid/malformed JSON
	body := `{"name": "John Doe", "age":`
	req, err := http.NewRequest("POST", "http://example.com/path", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Add("Content-Type", "application/json")

	// Call the function
	jsonData, reqJSON, err := RequestToJSON(req, "")
	require.NoError(t, err)

	// Check that the raw body is still captured as a string
	require.Equal(t, body, reqJSON.Body)

	// Check the JSON output
	require.Contains(t, string(jsonData), `"body":"{\"name\": \"John Doe\", \"age\":`)
}

func TestRequestToJSON_MultipartFormData(t *testing.T) {
	// Create a buffer for the multipart form data
	body := &bytes.Buffer{}
	writer := io.MultiWriter(body)
	_, err := writer.Write([]byte("----WebKitFormBoundary\nContent-Disposition: form-data; name=\"file\"; filename=\"test.txt\"\n\nfilecontent\n----WebKitFormBoundary--"))
	require.NoError(t, err)

	// Create a POST request with multipart/form-data
	req, err := http.NewRequest("POST", "http://example.com/upload", body)
	require.NoError(t, err)
	req.Header.Add("Content-Type", "multipart/form-data; boundary=----WebKitFormBoundary")

	// Call the function
	jsonData, reqJSON, err := RequestToJSON(req, "")
	require.NoError(t, err)

	// Verify that the body is captured as a raw string
	require.Contains(t, reqJSON.Body, "Content-Disposition: form-data; name=\"file\"; filename=\"test.txt\"")
	require.Contains(t, reqJSON.Body, "filecontent")

	// Check the JSON output
	require.Contains(t, string(jsonData), `"Content-Type":["multipart/form-data; boundary=----WebKitFormBoundary"]`)
}

func TestRequestToJSON_ErrorHandling(t *testing.T) {
	// Simulate a request with an unreadable body
	body := &readErrorCloser{}
	req, err := http.NewRequest("POST", "http://example.com/path", body)
	require.NoError(t, err)

	// Call the function
	jsonData, reqJSON, err := RequestToJSON(req, "")
	require.Error(t, err)

	// Ensure the function gracefully returns empty JSON and struct
	require.Equal(t, json.RawMessage{}, jsonData)
	require.Equal(t, RequestJSON{}, reqJSON)
}

// Helper to simulate a broken request body
type readErrorCloser struct{}

func (*readErrorCloser) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("read error")
}

func (*readErrorCloser) Close() error {
	return nil
}
