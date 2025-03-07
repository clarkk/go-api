package map_json

import (
	"bytes"
	"github.com/go-json-experiment/json"
)

type (
	Map 		[]Item
	Item struct {
		Key 	string
		Value 	any
	}
)

func (m Map) MarshalJSON() ([]byte, error){
	var b bytes.Buffer
	b.WriteString("{")
	for i, kv := range m {
		if i != 0 {
			b.WriteString(",")
		}
		//	Marshal key
		key, err := json.Marshal(kv.Key)
		if err != nil {
			return nil, err
		}
		b.Write(key)
		b.WriteString(":")
		//	Marshal value
		val, err := json.Marshal(kv.Value)
		if err != nil {
			return nil, err
		}
		b.Write(val)
	}
	b.WriteString("}")
	return b.Bytes(), nil
}