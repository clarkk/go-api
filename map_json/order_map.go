package map_json

import (
	"bytes"
	"encoding/json/v2"
	"encoding/json/jsontext"
)

type (
	Map struct {
		items	[]item
		index	map[string]int
	}
	
	item struct {
		key 	string
		value 	any
	}
)

func New() *Map {
	return &Map{
		index:	map[string]int{},
	}
}

func New_len(length int) *Map {
	return &Map{
		items:	make([]item, 0, length),
		index:	make(map[string]int, length),
	}
}

//	Set key/value entry
func (m *Map) Set(key string, value any){
	if i, ok := m.index[key]; ok {
		m.items[i].value = value
	} else {
		m.index[key] = len(m.items)
		m.items = append(m.items, item{
			key:	key,
			value:	value,
		})
	}
}

//	Get key/value entry
func (m *Map) Get(key string) (any, bool){
	if m == nil {
		return nil, false
	}
	if i, ok := m.index[key]; ok {
		return m.items[i].value, true
	}
	return nil, false
}

//	Get all keys in insertion order
func (m *Map) Keys() []string {
	if m == nil {
		return nil
	}
	keys := make([]string, len(m.items))
	for i, it := range m.items {
		keys[i] = it.key
	}
	return keys
}

//	Get number of entries
func (m *Map) Len() int {
	if m == nil {
		return 0
	}
	return len(m.items)
}

//	Marshals entries to JSON while preserving insertion order
func (m *Map) MarshalJSON() ([]byte, error){
	if m == nil {
		return []byte("null"), nil
	}
	
	var b bytes.Buffer
	enc := jsontext.NewEncoder(&b)
	
	if err := enc.WriteToken(jsontext.BeginObject); err != nil {
		return nil, err
	}
	
	for _, kv := range m.items {
		if err := enc.WriteToken(jsontext.String(kv.key)); err != nil {
			return nil, err
		}
		
		value_bytes, err := json.Marshal(kv.value)
		if err != nil {
			return nil, err
		}
		if err := enc.WriteValue(value_bytes); err != nil {
			return nil, err
		}
	}
	
	if err := enc.WriteToken(jsontext.EndObject); err != nil {
		return nil, err
	}
	
	return b.Bytes(), nil
}