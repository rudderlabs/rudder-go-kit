package sanitize

import (
	"bytes"
	stdjson "encoding/json"
	"io"
	"strconv"
)

func JSON(data []byte) ([]byte, error) {
	var (
		isKey    bool
		writePos int
		delims   []byte
		decoder  = stdjson.NewDecoder(bytes.NewReader(data))
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
			if v == '{' || v == '[' {
				delims = append(delims, byte(v))
			}
			if v == '{' {
				isKey = true // The next string token is a key
			}
			data[writePos] = byte(v)
			writePos++
			if v == '}' || v == ']' {
				delims = delims[:len(delims)-1]
				if decoder.More() { // nested objects or arrays
					data[writePos] = ','
					writePos++
				}
			}
		case string:
			writePos += copy(data[writePos:], strconv.Quote(v))
			if isKey {
				if ld() == '{' {
					data[writePos] = ':'
					writePos++
					isKey = false
				}
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
				if ld() == '{' {
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
				if ld() == '{' {
					isKey = true
				}
			}
		case nil:
			n := copy(data[writePos:], "null")
			writePos += n
			if decoder.More() {
				data[writePos] = ','
				writePos++
				if ld() == '{' {
					isKey = true
				}
			}
		}
	}

	// Return the compacted slice
	return data[:writePos], nil
}
