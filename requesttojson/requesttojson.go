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
	Query   map[string][]string `json:"query_parameters"`
}

func RequestToJSON(req *http.Request, replaceEmptyBodyWith string) (json.RawMessage, RequestJSON, error) {
	defer func() {
		_ = req.Body.Close()
	}()

	var bodyBytes []byte
	var err error

	// Read the body (if present)
	if req.Body != nil {
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return json.RawMessage{}, RequestJSON{}, err
		}
	}

	if len(bodyBytes) == 0 {
		bodyBytes = []byte(replaceEmptyBodyWith) // If body is empty, set it to an empty JSON object
	}

	// Create a RequestJSON struct to hold the necessary fields
	requestJSON := RequestJSON{
		Method:  req.Method,
		URL:     req.URL.RequestURI(),
		Proto:   req.Proto,
		Headers: req.Header,
		Query:   req.URL.Query(),
		Body:    string(bodyBytes),
	}

	req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Restore body for further reading

	requestBytes, err := json.Marshal(requestJSON)
	if err != nil {
		return json.RawMessage{}, RequestJSON{}, err
	}

	return requestBytes, requestJSON, nil
}
