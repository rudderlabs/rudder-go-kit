package stringify

import (
	"encoding/json"
	"fmt"
)

// Any converts any data to string
func Any(data any) string {
	if data == nil {
		return ""
	}
	switch d := data.(type) {
	case string:
		return d
	default:
		dataBytes, err := json.Marshal(d)
		if err != nil {
			return fmt.Sprint(d)
		}
		return string(dataBytes)
	}
}
