package errin

import (
	"strings"
	"github.com/clarkk/go-api/map_json"
)

type (
	Map			[]item[string]
	Map_lang	[]item[*Lang]
	
	Lang struct {
		Key		string
		Replace	Rep
	}
	Rep			map[string]any
	
	item[T any] struct {
		key		string
		value	T
	}
)

func (m *Map) Set(key, value string) {
	set(m, key, value)
}

func (m Map) Has(key string) bool {
	return has(m, key)
}

func (m Map) Output() *map_json.Map {
	return output(m)
}

func (m Map) String() string {
	if len(m) == 0 {
		return ""
	}
	s := make([]string, len(m))
	for i, v := range m {
		s[i] = v.key+": "+v.value
	}
	return strings.Join(s, ", ")
}

func (m *Map_lang) Set(key string, value *Lang) {
	set(m, key, value)
}

func (m Map_lang) Has(key string) bool {
	return has(m, key)
}

func (m Map_lang) Output() *map_json.Map {
	return output(m)
}

func (m Map_lang) String() string {
	if len(m) == 0 {
		return ""
	}
	s := make([]string, len(m))
	for i, v := range m {
		s[i] = v.key+": "+v.value.Key
	}
	return strings.Join(s, ", ")
}

func set[T any](l *[]item[T], key string, value T) {
	for i := range *l {
		if (*l)[i].key == key {
			(*l)[i].value = value
			return
		}
	}
	*l = append(*l, item[T]{
		key:	key,
		value:	value,
	})
}

func has[T any](l []item[T], key string) bool {
	for _, v := range l {
		if v.key == key {
			return true
		}
	}
	return false
}

func output[T any](l []item[T]) *map_json.Map {
	if len(l) == 0 {
		return nil
	}
	m := map_json.New_len(len(l))
	for _, v := range l {
		m.Set(v.key, v.value)
	}
	return m
}