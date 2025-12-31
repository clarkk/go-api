package map_json

import (
	"bytes"
	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
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
		index: map[string]int{},
	}
}

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

func (m *Map) Get(key string) (any, bool){
	if i, ok := m.index[key]; ok {
		return m.items[i].value, true
	}
	return nil, false
}

func (m *Map) Len() int {
	return len(m.items)
}

func (m *Map) MarshalJSON() ([]byte, error){
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