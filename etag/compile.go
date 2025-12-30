package etag

import (
	"fmt"
	"reflect"
	"strconv"
	"hash/crc32"
)

const null_string = "\x00"

var crc32_table = crc32.MakeTable(crc32.IEEE)

type etag struct {
	data []string
}

func New() *etag {
	return &etag{
		data: []string{},
	}
}

func (e *etag) Int(i int) *etag {
	return e.String(strconv.Itoa(i))
}

func (e *etag) Int_ptr(i *int) *etag {
	if i == nil {
		return e.String(null_string)
	} else {
		return e.String(strconv.Itoa(*i))
	}
}

func (e *etag) Int64(i int64) *etag {
	return e.String(strconv.FormatInt(i, 10))
}

func (e *etag) Int64_ptr(i *int64) *etag {
	if i == nil {
		return e.String(null_string)
	} else {
		return e.String(strconv.FormatInt(*i, 10))
	}
}

func (e *etag) Uint32(i uint32) *etag {
	return e.Uint64(uint64(i))
}

func (e *etag) Uint64(i uint64) *etag {
	return e.String(strconv.FormatUint(i, 10))
}

func (e *etag) Uint64_ptr(i *uint64) *etag {
	if i == nil {
		return e.String(null_string)
	} else {
		return e.String(strconv.FormatUint(*i, 10))
	}
}

func (e *etag) Float64(f float64) *etag {
	return e.String(strconv.FormatFloat(f, 'f', -1, 64))
}

func (e *etag) String(s string) *etag {
	e.data = append(e.data, s)
	return e
}

func (e *etag) String_ptr(s *string) *etag {
	if s == nil {
		return e.String(null_string)
	} else {
		return e.String(*s)
	}
}

func (e *etag) Bool(b bool) *etag {
	if b {
		return e.String("1")
	} else {
		return e.String("0")
	}
}

func (e *etag) Slice(slice any) *etag {
	if slice == nil {
		return e.String(null_string)
	}
	
	rv := reflect.ValueOf(slice)
	if rv.Kind() != reflect.Slice {
		panic("Slice: value is not a slice")
	}
	var sub_hash uint32
	for i := range rv.Len() {
		val := rv.Index(i).Interface()
		s := []byte(fmt.Sprint(val))
		sub_hash = hash_update(sub_hash, s, ",", i)
	}
	return e.Uint32(sub_hash)
}

func (e *etag) Compile() uint32 {
	var hash uint32
	for i, s := range e.data {
		hash = hash_update(hash, []byte(s), ":", i)
	}
	return hash
}

func hash_update(hash uint32, data []byte, separator string, index int) uint32 {
	if index > 0 {
		hash = crc32.Update(hash, crc32_table, []byte(separator))
	}
	return crc32.Update(hash, crc32_table, data)
}