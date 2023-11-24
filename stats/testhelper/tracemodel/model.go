package tracemodel

import "time"

type Span struct {
	Name                   string                 `json:"Name"`
	SpanContext            SpanContext            `json:"SpanContext"`
	Parent                 Parent                 `json:"Parent"`
	SpanKind               int                    `json:"SpanKind"`
	StartTime              time.Time              `json:"StartTime"`
	EndTime                time.Time              `json:"EndTime"`
	Attributes             []Attributes           `json:"Attributes"`
	Events                 any                    `json:"Events"`
	Links                  any                    `json:"Links"`
	Status                 Status                 `json:"Status"`
	DroppedAttributes      int                    `json:"DroppedAttributes"`
	DroppedEvents          int                    `json:"DroppedEvents"`
	DroppedLinks           int                    `json:"DroppedLinks"`
	ChildSpanCount         int                    `json:"ChildSpanCount"`
	Resource               []Resource             `json:"Resource"`
	InstrumentationLibrary InstrumentationLibrary `json:"InstrumentationLibrary"`
}
type SpanContext struct {
	TraceID    string `json:"TraceID"`
	SpanID     string `json:"SpanID"`
	TraceFlags string `json:"TraceFlags"`
	TraceState string `json:"TraceState"`
	Remote     bool   `json:"Remote"`
}
type Parent struct {
	TraceID    string `json:"TraceID"`
	SpanID     string `json:"SpanID"`
	TraceFlags string `json:"TraceFlags"`
	TraceState string `json:"TraceState"`
	Remote     bool   `json:"Remote"`
}
type Value struct {
	Type  string `json:"Type"`
	Value string `json:"Value"`
}
type Attributes struct {
	Key   string `json:"Key"`
	Value Value  `json:"Value"`
}
type Status struct {
	Code        string `json:"Code"`
	Description string `json:"Description"`
}
type Resource struct {
	Key   string `json:"Key"`
	Value Value  `json:"Value"`
}
type InstrumentationLibrary struct {
	Name      string `json:"Name"`
	Version   string `json:"Version"`
	SchemaURL string `json:"SchemaURL"`
}
