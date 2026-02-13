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
	for i := range *m {
		if (*m)[i].key == key {
			(*m)[i].value = value
			return
		}
	}
	*m = append(*m, item[string]{
		key:	key,
		value:	value,
	})
}

func (m *Map_lang) Set(key string, value *Lang) {
	for i := range *m {
		if (*m)[i].key == key {
			(*m)[i].value = value
			return
		}
	}
	*m = append(*m, item[*Lang]{
		key:	key,
		value:	value,
	})
}

func (m Map) Has(key string) bool {
	for _, v := range m {
		if v.key == key {
			return true
		}
	}
	return false
}

func (m Map_lang) Has(key string) bool {
	for _, v := range m {
		if v.key == key {
			return true
		}
	}
	return false
}

func (m Map) Output() *map_json.Map {
	if len(m) == 0 {
		return nil
	}
	res := map_json.New_len(len(m))
	for _, v := range m {
		res.Set(v.key, v.value)
	}
	return res
}

func (m Map_lang) Output() *map_json.Map {
	if len(m) == 0 {
		return nil
	}
	res := map_json.New_len(len(m))
	for _, v := range m {
		res.Set(v.key, v.value)
	}
	return res
}

func (m Map) String() string {
	if m == nil || len(m) == 0 {
		return ""
	}
	s := make([]string, len(m))
	for k, v := range m {
		s[k] = v.key+": "+v.value
	}
	return strings.Join(s, ", ")
}

func (m Map_lang) String() string {
	if m == nil || len(m) == 0 {
		return ""
	}
	s := make([]string, len(m))
	for k, v := range m {
		s[k] = v.key+": "+v.value.Key
	}
	return strings.Join(s, ", ")
}