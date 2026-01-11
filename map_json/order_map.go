package map_json

import (
	"bytes"
	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
)

type (
	Map[V any] struct {
		items	[]item[V]
		index	map[string]int
	}
	
	item[V any] struct {
		key 	string
		value 	V
	}
)

func New[V any]() *Map[V] {
	return &Map[V]{
		index: map[string]int{},
	}
}

func (m *Map[V]) Set(key string, value V){
	if i, ok := m.index[key]; ok {
		m.items[i].value = value
	} else {
		m.index[key] = len(m.items)
		m.items = append(m.items, item[V]{
			key:	key,
			value:	value,
		})
	}
}

func (m *Map[V]) Get(key string) (V, bool){
	if i, ok := m.index[key]; ok {
		return m.items[i].value, true
	}
	var zero V
	return zero, false
}

func (m *Map[V]) Len() int {
	return len(m.items)
}

func (m *Map[V]) MarshalJSON() ([]byte, error){
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