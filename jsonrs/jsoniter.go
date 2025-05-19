package jsonrs

import (
	"io"

	jsoniter "github.com/json-iterator/go"
)

var defaultJsoniter = jsoniter.ConfigCompatibleWithStandardLibrary

// jsoniterJSON is the JSON implementation of github.com/json-iterator/go.
type jsoniterJSON struct{}

func (j *jsoniterJSON) Marshal(v any) ([]byte, error) {
	return defaultJsoniter.Marshal(v)
}

func (j *jsoniterJSON) MarshalIndent(v any, prefix, indent string) ([]byte, error) {
	return defaultJsoniter.MarshalIndent(v, prefix, indent)
}

func (j *jsoniterJSON) Unmarshal(data []byte, v any) error {
	return defaultJsoniter.Unmarshal(data, v)
}

func (j *jsoniterJSON) MarshalToString(v any) (string, error) {
	return defaultJsoniter.MarshalToString(v)
}

func (j *jsoniterJSON) NewDecoder(r io.Reader) Decoder {
	return defaultJsoniter.NewDecoder(r)
}

func (j *jsoniterJSON) NewEncoder(w io.Writer) Encoder {
	return defaultJsoniter.NewEncoder(w)
}

func (j *jsoniterJSON) Valid(data []byte) bool {
	return defaultJsoniter.Valid(data)
}
