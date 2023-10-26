package idem

import (
	"fmt"
	"context"
	"net/http"
	"github.com/go-errors/errors"
	"github.com/go-json-experiment/json"
	"github.com/clarkk/go-util/rdb"
)

const (
	IDEM_HEADER_KEY 	= "Idempotency-Key"
	IDEM_HEADER_LENGTH 	= 40
	IDEM_EXPIRES 		= 60 * 60 * 24
	IDEM_HASH 			= "API-IDEM:%s:%s"
)

type (
	Idempotency struct {
		error 		error
		hash 		string
		cached 		bool
		http_code 	int
		res 		string
	}
	
	cache struct {
		Http_code 	int 	`redis:"http_code"`
		Res 		string 	`redis:"res"`
	}
)

func (l *Idempotency) Cache() (bool, int, string) {
	if l.cached {
		return l.cached, l.http_code, l.res
	}
	return false, 0, ""
}

func (l *Idempotency) Error() (int, error) {
	if l.error == nil {
		return 0, nil
	}
	return l.http_code, l.error
}

func (l *Idempotency) Store_JSON(http_code int, res map[string]interface{}){
	//	Only store HTTP 200 OK
	if http_code != http.StatusOK {
		return
	}
	
	b, err := json.Marshal(res)
	if err != nil {
		panic("Idempotency store JSON encode: "+err.Error())
	}
	l.Store(http_code, string(b))
}

func (l *Idempotency) Store(http_code int, res string){
	//	Only store HTTP 200 OK
	if http_code != http.StatusOK {
		return
	}
	
	if err := rdb.Hset(context.Background(), l.hash, cache{
		Http_code:	http_code,
		Res:		res,
	}, IDEM_EXPIRES); err != nil {
		panic("Could not store idempotency response: "+err.Error())
	}
}

func Init(r *http.Request, uid string) *Idempotency {
	if !rdb.Connected() {
		panic("Redis is not connected")
	}
	
	l := &Idempotency{}
	
	//	Check if header is provided
	key := r.Header.Get(IDEM_HEADER_KEY)
	if key == "" {
		l.error 		= errors.New(fmt.Sprintf("%s header must be provided", IDEM_HEADER_KEY))
		l.http_code 	= http.StatusNotAcceptable
		return l
	}
	
	//	Check if value has the right length
	if len(key) > IDEM_HEADER_LENGTH {
		l.error 		= errors.New(fmt.Sprintf("%s header value must not be longer than %d chars", IDEM_HEADER_KEY, IDEM_HEADER_LENGTH))
		l.http_code 	= http.StatusNotAcceptable
		return l
	}
	
	//	Check if request is duplicate
	var ref cache
	l.hash = fmt.Sprintf(IDEM_HASH, uid, key)
	rdb.Hgetall(r.Context(), l.hash, &ref)
	if ref.Http_code != 0 {
		l.cached 	= true
		l.http_code = ref.Http_code
		l.res 		= ref.Res
	}
	return l
}