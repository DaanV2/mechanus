package xformat

import (
	"encoding"
	"encoding/json"
)

// Unmarshal tries to check if result matches any of the following marshallers, and if that fails, unmarshals via json
// [encoding.TextUnmarshaler] -> [encoding.BinaryUnmarshaler] -> [json.Unmarshal]
func Unmarshal(data []byte, result any) error {
	if v, ok := result.(encoding.TextUnmarshaler); ok {
		return v.UnmarshalText(data)
	}

	if v, ok := result.(encoding.BinaryUnmarshaler); ok {
		return v.UnmarshalBinary(data)
	}

	return json.Unmarshal(data, result)
}

func Marshal(item any) ([]byte, error) {
	if v, ok := item.(encoding.TextMarshaler); ok {
		return v.MarshalText()
	}

	if v, ok := item.(encoding.BinaryMarshaler); ok {
		return v.MarshalBinary()
	}

	return json.Marshal(item)
}
