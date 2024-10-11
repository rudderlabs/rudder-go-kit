package requesttojson

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

// Struct to represent the HTTP request in JSON format
type RequestJSON struct {
	Method  string              `json:"method"`
	URL     string              `json:"url"`
	Proto   string              `json:"proto"`
	Headers map[string][]string `json:"headers"`
	Body    string              `json:"body"`
	Query   map[string][]string `json:"query"`
}

func RequestToJSON(req *http.Request) (json.RawMessage, RequestJSON) {
	// Parse query parameters from URL
	queryParams := req.URL.Query()

	// Create a RequestJSON struct to hold the necessary fields
	requestJSON := RequestJSON{
		Method:  req.Method,
		URL:     req.URL.RequestURI(),
		Proto:   req.Proto,
		Headers: req.Header,
		Query:   queryParams, // Include query parameters here
	}

	// Read the body (if present)
	if req.Body != nil {
		bodyBytes, err := io.ReadAll(req.Body)
		if err != nil {
			return json.RawMessage{}, RequestJSON{}
		}
		requestJSON.Body = string(bodyBytes)
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Restore body for further reading
	}

	// Create a buffer to store the JSON output
	var buf bytes.Buffer

	// Create a JSON encoder and disable HTML escaping
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false) // Disable escaping of special characters like & and +

	// Marshal the struct into JSON using the custom encoder
	err := encoder.Encode(requestJSON)
	if err != nil {
		return json.RawMessage{}, RequestJSON{}
	}

	return json.RawMessage(buf.Bytes()), requestJSON
}
