package jsonrs

import (
	"github.com/tidwall/gjson"
)

type tidwallJSON struct{}

func (j *tidwallJSON) Valid(data []byte) bool {
	return gjson.ValidBytes(data)
}
