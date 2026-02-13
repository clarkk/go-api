package errin

import (
	"strings"
	"github.com/clarkk/go-api/map_json"
)

type (
	Map struct {
		base
	}
	
	Map_lang struct {
		base
	}
	
	Lang struct {
		Key		string
		Replace	Rep
	}
	Rep			map[string]any
	
	base struct {
		*map_json.Map
	}
)

func (b *base) init(){
	if b.Map == nil {
		b.Map = map_json.New()
	}
}

func (b base) Has(key string) bool {
	if b.Map == nil {
		return false
	}
	_, ok := b.Get(key)
	return ok
}

func (m *Map) Set(key, msg string) {
	m.init()
	m.Map.Set(key, msg)
}

func (m *Map_lang) Set(key string, lang *Lang) {
	m.init()
	m.Map.Set(key, lang)
}

func (m Map_lang) String() string {
	if m.empty() {
		return ""
	}
	keys := m.Keys()
	s := make([]string, len(keys))
	for i, k := range keys {
		val, _ := m.Get(k)
		s[i] = k + ": " + val.(*Lang).Key
	}
	return strings.Join(s, ", ")
}

func (m Map) String() string {
	if m.empty() {
		return ""
	}
	keys := m.Keys()
	s := make([]string, len(keys))
	for i, k := range keys {
		val, _ := m.Get(k)
		s[i] = k + ": " + val.(string)
	}
	return strings.Join(s, ", ")
}

func (b base) empty() bool {
	return b.Map == nil || b.Len() == 0
}