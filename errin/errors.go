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
		nested	*Map
	}
	
	item_lang struct {
		key		string
		value	*Lang
		nested	*Map_lang
	}
)

func (m *Map) Set(key, value string){
	m.set(item{
		key:	key,
		value:	value,
	})
}

func (m *Map) Set_map(key string, value Map){
	m.set(item{
		key:	key,
		nested:	&value,
	})
}

func (m *Map_lang) Set(key string, value *Lang){
	m.set(item_lang{
		key:	key,
		value:	value,
	})
}

func (m *Map_lang) Set_map(key string, value Map_lang){
	m.set(item_lang{
		key:	key,
		nested:	&value,
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
		if v.nested != nil {
			res.Set(v.key, v.nested.Output())
		} else {
			res.Set(v.key, v.value)
		}
	}
	return res
}

func (m Map_lang) Output() *map_json.Map {
	if len(m) == 0 {
		return nil
	}
	res := map_json.New_len(len(m))
	for _, v := range m {
		if v.nested != nil {
			res.Set(v.key, v.nested.Output())
		} else {
			res.Set(v.key, v.value)
		}
	}
	return res
}

func (m Map) String() string {
	s := make([]string, len(m))
	for i, v := range m {
		if v.nested != nil {
			s[i] = v.key+": {"+v.nested.String()+"}"
		} else {
			s[i] = v.key+": "+v.value
		}
	}
	return strings.Join(s, ", ")
}

func (m Map_lang) String() string {
	s := make([]string, len(m))
	for i, v := range m {
		if v.nested != nil {
			s[i] = v.key+": {"+v.nested.String()+"}"
		} else if v.value != nil {
			s[i] = v.key+": "+v.value.Key
		} else {
			s[i] = v.key+": <nil>"
		}
	}
	return strings.Join(s, ", ")
}

func (m *Map) set(value item){
	for i := range *m {
		if (*m)[i].key == value.key {
			(*m)[i] = value
			return
		}
	}
	*m = append(*m, value)
}

func (m *Map_lang) set(value item_lang){
	for i := range *m {
		if (*m)[i].key == value.key {
			(*m)[i] = value
			return
		}
	}
	*m = append(*m, value)
}