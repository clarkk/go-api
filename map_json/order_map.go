package map_json

import (
	"bytes"
	"github.com/go-json-experiment/json"
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

func (m *Map) MarshalJSON() ([]byte, error){
	var b bytes.Buffer
	b.WriteString("{")
	for i, kv := range m.items {
		if i != 0 {
			b.WriteString(",")
		}
		//	Marshal key
		key, err := json.Marshal(kv.key)
		if err != nil {
			return nil, err
		}
		b.Write(key)
		b.WriteString(":")
		//	Marshal value
		val, err := json.Marshal(kv.value)
		if err != nil {
			return nil, err
		}
		b.Write(val)
	}
	b.WriteString("}")
	return b.Bytes(), nil
}