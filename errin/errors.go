package errin

import (
	"strings"
	"github.com/clarkk/go-api/map_json"
)

type (
	Map			list[string]
	Map_lang	list[*Lang]
	
	Lang struct {
		Key		string
		Replace	Rep
	}
	Rep			map[string]any
	
	list[T any]	[]item[T]
	
	item[T any] struct {
		key		string
		value	T
	}
)

func (l *list[T]) Set(key string, value T){
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

func (l list[T]) Has(key string) bool {
	for _, v := range l {
		if v.key == key {
			return true
		}
	}
	return false
}

func (l list[T]) Map() *map_json.Map {
	if len(l) == 0 {
		return nil
	}
	m := map_json.New_len(len(l))
	for _, v := range l {
		m.Set(v.key, v.value)
	}
	return m
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