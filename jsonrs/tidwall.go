package jsonrs

import (
	"github.com/tidwall/gjson"
)

type tidwallValidator struct{}

func (j *tidwallValidator) Valid(data []byte) bool {
	return gjson.ValidBytes(data)
}
