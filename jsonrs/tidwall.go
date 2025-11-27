package jsonrs

import (
	"io"

	"github.com/tidwall/gjson"
)

type tidwallJSON struct {
	stdJSON
}

func (j *tidwallJSON) Valid(data []byte) bool {
	return gjson.ValidBytes(data)
}

func (j *tidwallJSON) NewDecoder(r io.Reader) Decoder {
	// tidwall is only for validation, so delegate decoding to stdlib
	return j.stdJSON.NewDecoder(r)
}

func (j *tidwallJSON) NewEncoder(w io.Writer) Encoder {
	// tidwall is only for validation, so delegate encoding to stdlib
	return j.stdJSON.NewEncoder(w)
}
