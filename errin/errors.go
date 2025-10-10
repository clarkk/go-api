package errin

import (
	"errors"
	"strings"
)

type (
	Map			map[string]error
	Map_lang	map[string]*Lang
	
	Lang struct {
		Key		string
		Replace	Rep
	}
	Rep			map[string]any
)

func (m *Map) Set(key, msg string){
	if *m == nil {
		*m = Map{}
	}
	(*m)[key] = errors.New(msg)
}

func (m Map) Has(key string) bool {
	_, ok := m[key]
	return ok
}

func (m *Map_lang) Set(key string, lang *Lang){
	if *m == nil {
		*m = Map_lang{}
	}
	(*m)[key] = lang
}

func (m Map_lang) Has(key string) bool {
	_, ok := m[key]
	return ok
}

func (m *Map_lang) String() string {
	if m == nil {
		return ""
	}
	
	s := make([]slice, len(*m))
	var i int
	for k, v := *m {
		s[i] = k+": "+v.Key
		i++
	}
	return strings.Join(s, ", ")
}