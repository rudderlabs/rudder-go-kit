package tracemodel

type ZipkinTrace struct {
	TraceID     string                  `json:"traceId"`
	ParentID    string                  `json:"parentId"`
	ID          string                  `json:"id"`
	Name        string                  `json:"name"`
	Timestamp   int64                   `json:"timestamp"`
	Duration    int64                   `json:"duration"`
	Tags        map[string]string       `json:"tags"`
	Annotations []ZipkinTraceAnnotation `json:"annotations"`
}

type ZipkinTraceAnnotation struct {
	Timestamp int64  `json:"timestamp"`
	Value     string `json:"value"`
}
