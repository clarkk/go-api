package etag

import (
	"strings"
	"strconv"
	"net/http"
	"hash/crc32"
	"github.com/clarkk/go-api"
)

type (
	Matcher interface {
		Match_etag(id uint64, etag_header string) bool
	}
	
	etag struct {
		data []string
	}
)

func Match(a *api.Request, id uint64, m Matcher) bool {
	if id == 0 {
		return true
	}
	etag_header := strip_encapsulation(a.Request_header("If-Match"))
	if etag_header == "" {
		return true
	}
	if !m.Match_etag(id, etag_header) {
		a.Errorf(http.StatusPreconditionFailed, "Entity-tag mismatch")
		return false
	}
	return true
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

func (e *etag) Uint64(i uint64) *etag {
	e.data = append(e.data, strconv.FormatUint(i, 10))
	return e
}

func (e *etag) Uint64_ptr(i *uint64) *etag {
	if i == nil {
		e.data = append(e.data, "")
	} else {
		e.data = append(e.data, strconv.FormatUint(*i, 10))
	}
	return e
}

func (e *etag) String(s string) *etag {
	e.data = append(e.data, s)
	return e
}

func (e *etag) Compile() uint32 {
	crc32q := crc32.MakeTable(0xedb88320)
	s := strings.Join(e.data, ":")
	return crc32.Checksum([]byte(s), crc32q)
}

func strip_encapsulation(s string) string {
	return strings.Trim(s, `"`)
}