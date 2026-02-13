package errin

import (
	"strings"
	"github.com/clarkk/go-api/map_json"
)

type (
	Map			[]item
	Map_lang	[]item_lang
	
	Lang struct {
		Key		string
		Replace	Rep
	}
	Rep			map[string]any
	
	item struct {
		key		string
		value	string
	}
	
	item_lang struct {
		key		string
		value	*Lang
	}
)

func (m *Map) Set(key, value string){
	for i := range *m {
		if (*m)[i].key == key {
			(*m)[i].value = value
			return
		}
	}
	*m = append(*m, item{
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
	*m = append(*m, item_lang{
		key:	key,
		value:	value,
	})
}

func (m Map) Get(key string) (string, bool){
	for _, v := range m {
		if v.key == key {
			return v.value, true
		}
	}
	return "", false
}

func (m Map_lang) Get(key string) (*Lang, bool) {
	for _, v := range m {
		if v.key == key {
			return v.value, true
		}
	}
	return nil, false
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