package sanitize

import (
	"bytes"
	stdjson "encoding/json"
	"io"
	"strconv"

	"github.com/goccy/go-json"
)

func JSON(data []byte) ([]byte, error) {
	var (
		wroteKey bool
		writePos int
		delims   []byte
		decoder  = json.NewDecoder(bytes.NewReader(data))
	)

	ld := func() byte { // last delimiter
		if len(delims) == 0 {
			return 0
		}
		return delims[len(delims)-1]
	}

	for {
		token, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		switch v := token.(type) {
		case stdjson.Delim:
			data[writePos] = byte(v)
			writePos++
			if v == '{' || v == '[' {
				delims = append(delims, byte(v))
				if v == '{' {
					wroteKey = false
				}
			} else if v == '}' || v == ']' {
				wroteKey = false
				delims = delims[:len(delims)-1]
				if decoder.More() { // nested objects or arrays
					data[writePos] = ','
					writePos++
				}
			}
		case string:
			writePos += copy(data[writePos:], strconv.Quote(v))
			lastDelim := ld()
			if lastDelim == '{' && !wroteKey {
				data[writePos] = ':'
				writePos++
				wroteKey = true
			} else if decoder.More() {
				data[writePos] = ','
				writePos++
				wroteKey = false
			}
		case float64:
			n := copy(data[writePos:], strconv.FormatFloat(v, 'f', -1, 64))
			writePos += n
			wroteKey = false
			if decoder.More() {
				data[writePos] = ','
				writePos++
			}
		case bool:
			if v {
				writePos += copy(data[writePos:], "true")
			} else {
				writePos += copy(data[writePos:], "false")
			}
			wroteKey = false
			if decoder.More() {
				data[writePos] = ','
				writePos++
			}
		case nil:
			writePos += copy(data[writePos:], "null")
			wroteKey = false
			if decoder.More() {
				data[writePos] = ','
				writePos++
			}
		}
	}

	// Return the compacted slice
	return data[:writePos], nil
}
