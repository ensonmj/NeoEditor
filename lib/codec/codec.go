package codec

import (
	"bytes"
	"encoding/json"
	"errors"
)

func Serialize(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func Deserialize(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

type Envelope struct {
	Method    string
	Arguments interface{}
}

// For deserialize Arguments in Envelope
// It implements Marshaler and Unmarshaler and can be used to delay
// JSON decoding or precompute a JSON encoding.
type RawMessage []byte

func (m *RawMessage) MarshalJSON() ([]byte, error) {
	return *m, nil
}

func (m *RawMessage) UnmarshalJSON(data []byte) error {
	if m == nil {
		return errors.New("codec.RawMessage: UnmarshalJSON on nil pointer")
	}
	*m = append((*m)[0:0], data...)
	return nil
}

type KeyVal struct {
	Key string
	Val interface{}
}

// Define an ordered map
// dict := map[string]interface{}{
//     "orderedMap": OrderedMap{
//			{"name", "John"},
//			{"age", 20},
//		}
// }
type OrderedMap []KeyVal

// Implement the json.Marshaler interface
func (omap OrderedMap) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer

	buf.WriteString("{")
	for i, kv := range omap {
		if i != 0 {
			buf.WriteString(",")
		}

		// marshal key
		key, err := json.Marshal(kv.Key)
		if err != nil {
			return nil, err
		}
		buf.Write(key)
		buf.WriteString(":")
		// marshal value
		val, err := json.Marshal(kv.Val)
		if err != nil {
			return nil, err
		}
		buf.Write(val)
	}

	buf.WriteString("}")
	return buf.Bytes(), nil
}
