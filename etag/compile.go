package etag

import (
	"strings"
	"strconv"
	"hash/crc32"
)

type etag struct {
	data []string
}

func New() *etag {
	return &etag{
		data: []string{},
	}
}

func (e *etag) Int(i int) *etag {
	e.data = append(e.data, strconv.Itoa(i))
	return e
}

func (e *etag) Int_ptr(i *int) *etag {
	if i == nil {
		e.data = append(e.data, "\x00")
	} else {
		e.data = append(e.data, strconv.Itoa(*i))
	}
	return e
}

func (e *etag) Int64(i int64) *etag {
	e.data = append(e.data, strconv.FormatInt(i, 10))
	return e
}

func (e *etag) Int64_ptr(i *int64) *etag {
	if i == nil {
		e.data = append(e.data, "\x00")
	} else {
		e.data = append(e.data, strconv.FormatInt(*i, 10))
	}
	return e
}

func (e *etag) Uint32(i uint32) *etag {
	return e.Uint64(uint64(i))
}

func (e *etag) Uint64(i uint64) *etag {
	e.data = append(e.data, strconv.FormatUint(i, 10))
	return e
}

func (e *etag) Uint64_ptr(i *uint64) *etag {
	if i == nil {
		e.data = append(e.data, "\x00")
	} else {
		e.data = append(e.data, strconv.FormatUint(*i, 10))
	}
	return e
}

func (e *etag) Float64(f float64) *etag {
	e.data = append(e.data, strconv.FormatFloat(f, 'f', -1, 64))
	return e
}

func (e *etag) String(s string) *etag {
	e.data = append(e.data, s)
	return e
}

func (e *etag) String_ptr(s *string) *etag {
	if s == nil {
		e.data = append(e.data, "\x00")
	} else {
		e.data = append(e.data, *s)
	}
	return e
}

func (e *etag) Compile() uint32 {
	crc32q := crc32.MakeTable(0xedb88320)
	s := strings.Join(e.data, ":")
	return crc32.Checksum([]byte(s), crc32q)
}