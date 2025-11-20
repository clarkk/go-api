package etag

import (
	"bytes"
	"strings"
	"strconv"
	"hash/crc32"
	"encoding/gob"
)

const null_string = "\x00"

type etag struct {
	data []string
}

func New() *etag {
	return &etag{
		data: []string{},
	}
}

func (e *etag) Int(i int) *etag {
	e.String(strconv.Itoa(i))
	return e
}

func (e *etag) Int_ptr(i *int) *etag {
	if i == nil {
		e.String(null_string)
	} else {
		e.String(strconv.Itoa(*i))
	}
	return e
}

func (e *etag) Int64(i int64) *etag {
	e.String(strconv.FormatInt(i, 10))
	return e
}

func (e *etag) Int64_ptr(i *int64) *etag {
	if i == nil {
		e.String(null_string)
	} else {
		e.String(strconv.FormatInt(*i, 10))
	}
	return e
}

func (e *etag) Uint32(i uint32) *etag {
	return e.Uint64(uint64(i))
}

func (e *etag) Uint64(i uint64) *etag {
	e.String(strconv.FormatUint(i, 10))
	return e
}

func (e *etag) Uint64_ptr(i *uint64) *etag {
	if i == nil {
		e.String(null_string)
	} else {
		e.String(strconv.FormatUint(*i, 10))
	}
	return e
}

func (e *etag) Float64(f float64) *etag {
	e.String(strconv.FormatFloat(f, 'f', -1, 64))
	return e
}

func (e *etag) String(s string) *etag {
	e.data = append(e.data, s)
	return e
}

func (e *etag) String_ptr(s *string) *etag {
	if s == nil {
		e.String(null_string)
	} else {
		e.String(*s)
	}
	return e
}

func (e *etag) Bool(b bool) *etag {
	if b {
		e.String("1")
	} else {
		e.String("0")
	}
	return e
}

func (e *etag) Slice(values []any) *etag {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	for _, v := range values {
		if err := enc.Encode(v); err != nil {
			panic(err)
		}
	}
	e.String(buf.String())
	return e
}

func (e *etag) Compile() uint32 {
	crc32q := crc32.MakeTable(0xedb88320)
	s := strings.Join(e.data, ":")
	return crc32.Checksum([]byte(s), crc32q)
}