package etag

import (
	"fmt"
	"strings"
	"net/http"
	"github.com/clarkk/go-api"
)

const HEADER_IF_MATCH = "If-Match"

type Matcher interface {
	Match_etag(id uint64, etag_header string) (bool, bool, error)
}

//	Match "If-Match" matches entity-tag of entry id
func Match(a *api.Request, id uint64, m Matcher) (int, error){
	if id == 0 {
		return 0, nil
	}
	etag_header := strip_encapsulation(a.Request_header(HEADER_IF_MATCH))
	if etag_header == "" {
		return 0, nil
	}
	found, match, err := m.Match_etag(id, etag_header)
	if err != nil {
		return 0, err
	}
	if !found {
		return http.StatusNotFound, nil
	}
	if !match {
		return http.StatusPreconditionFailed, fmt.Errorf("Entity-tag mismatch")
	}
	return 0, nil
}

//	Disallow "If-Match"
func Disallow_match(a *api.Request) bool {
	if strip_encapsulation(a.Request_header(HEADER_IF_MATCH)) != "" {
		a.Errorf(http.StatusPreconditionFailed, "%s header not allowed", HEADER_IF_MATCH)
		return false
	}
	return true
}

func strip_encapsulation(s string) string {
	return strings.Trim(s, `"`)
}