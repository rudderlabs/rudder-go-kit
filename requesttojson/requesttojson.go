package requesttojson

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

// RequestJSON represents an HTTP request in a simplified JSON-friendly format.
type RequestJSON struct {
	Method  string              `json:"method"`
	URL     string              `json:"url"`
	Proto   string              `json:"proto"`
	Headers map[string][]string `json:"headers"`
	Body    string              `json:"body"`
	Query   map[string][]string `json:"query_parameters"`
}

// RequestToJson converts a http.Request into a JSON-friendly RequestJson struct and returns its pointer
// Function takes 2 arguments
//   - req *http.Request : The request that needs to be converted to RequestJson
//   - defaultBodyContent string: RequestJson's Body field is assigned with defaultBodyContent in case of a request with an empty body
func RequestToJSON(req *http.Request, defaultBodyContent string) (*RequestJSON, error) {
	defer func() {
		if req.Body != nil {
			_ = req.Body.Close()
		}
	}()
	var bodyBytes []byte
	var err error

	// Read the body (if present)
	if req.Body != nil {
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("reading request body: %w", err)
		}
	}
	if len(bodyBytes) == 0 {
		bodyBytes = []byte(defaultBodyContent) // If body is empty, set it to an empty JSON object
	}

	// Restore body for further reading
	req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// Create a RequestJSON struct to hold the necessary fields
	requestJSON := &RequestJSON{
		Method:  req.Method,
		URL:     req.URL.RequestURI(),
		Proto:   req.Proto,
		Headers: req.Header,
		Query:   req.URL.Query(),
		Body:    string(bodyBytes),
	}
	return requestJSON, nil
}
