package errin

import (
	"fmt"
	"strings"
	"github.com/clarkk/go-api/map_json"
)

type (
	Map struct {
		*map_json.Map
	}

	Map_lang struct {
		*map_json.Map
	}
	
	Lang struct {
		Key		string
		Replace	Rep
	}
	Rep			map[string]any
)

func (m *Map) Set(key, msg string){
	if m.Map == nil {
		m.Map = map_json.New()
	}
	m.Map.Set(key, msg)
}

func (m Map) Has(key string) bool {
	if m.Map == nil {
		return false
	}
	_, ok := m.Map.Get(key)
	return ok
}

func (m *Map_lang) Set(key string, lang *Lang){
	if m.Map == nil {
		m.Map = map_json.New()
	}
	m.Map.Set(key, lang)
}

func (m Map_lang) Has(key string) bool {
	if m.Map == nil {
		return false
	}
	_, ok := m.Map.Get(key)
	return ok
}

func (m *Map_lang) String() string {
	if m == nil || m.Map == nil {
		return ""
	}
	keys := m.Map.Keys()
	s := make([]string, len(keys))
	for i, k := range keys {
		val, _ := m.Map.Get(k)
		s[i] = k+": "+val.(*Lang).Key
	}
	return strings.Join(s, ", ")
}

func (m *Map) String() string {
	if m == nil || m.Map == nil {
		return ""
	}
	keys := m.Map.Keys()
	s := make([]string, len(keys))
	for i, k := range keys {
		val, _ := m.Map.Get(k)
		s[i] = k+": "+val.(string)
	}
	return strings.Join(s, ", ")
}