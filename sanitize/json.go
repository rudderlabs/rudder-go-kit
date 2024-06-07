package sanitize

import (
	"bytes"
	stdjson "encoding/json"
	"io"
	"strconv"
)

func JSON(data []byte) ([]byte, error) {
	var (
		isKey              bool
		writePos, inObject int

		decoder = stdjson.NewDecoder(bytes.NewReader(data))
	)

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
			if v == '{' {
				isKey = true // The next string token is a key
				inObject++
			} else if v == '}' {
				inObject--
			}
			data[writePos] = byte(v)
			writePos++
			if (v == '}' || v == ']') && decoder.More() { // nested objects or arrays
				data[writePos] = ','
				writePos++
			}
		case string:
			writePos += copy(data[writePos:], strconv.Quote(v))
			if inObject == 0 {
				continue
			}
			if isKey {
				data[writePos] = ':'
				writePos++
				isKey = false
			} else if decoder.More() {
				data[writePos] = ','
				writePos++
				isKey = true
			}
		case float64:
			n := copy(data[writePos:], strconv.FormatFloat(v, 'f', -1, 64))
			writePos += n
			if decoder.More() {
				data[writePos] = ','
				writePos++
				if inObject > 0 {
					isKey = true
				}
			}
		case bool:
			if v {
				n := copy(data[writePos:], "true")
				writePos += n
			} else {
				n := copy(data[writePos:], "false")
				writePos += n
			}
			if decoder.More() {
				data[writePos] = ','
				writePos++
				if inObject > 0 {
					isKey = true
				}
			}
		case nil:
			n := copy(data[writePos:], "null")
			writePos += n
			if decoder.More() {
				data[writePos] = ','
				writePos++
				if inObject > 0 {
					isKey = true
				}
			}
		}
	}

	// Return the compacted slice
	return data[:writePos], nil
}
