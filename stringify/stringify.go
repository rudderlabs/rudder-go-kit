package stringify

import (
	"encoding/json"
	"fmt"
)

func Data(data any) string {
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
