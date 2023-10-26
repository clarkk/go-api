package api

import (
	"fmt"
	"context"
	"net/http"
	"github.com/clarkk/go-util/rdb"
)

const (
	IDEM_HEADER_KEY 	= "X-Idempotency-Key"
	IDEM_HEADER_LENGTH 	= 40
	IDEM_EXPIRES 		= 60 * 60 * 24
	IDEM_HASH 			= "API-IDEM:%s:%s"
)

type (
	idem_local struct {
		error 		string
		hash 		string
		http_code 	int
		res 		string
	}
	
	idem_remote struct {
		Http_code 	int 	`redis:"http_code"`
		Res 		string 	`redis:"res"`
	}
)

func (l *idem_local) Cache() (int, string) {
	if l.res == "" {
		return 0, ""
	}
	return l.http_code, l.res
}

func (l *idem_local) Error() (int, string) {
	if l.error == "" {
		return 0, ""
	}
	return l.http_code, l.error
}

func (l *idem_local) Store(http_code int, res string){
	if err := rdb.Hset(context.Background(), l.hash, idem_remote{
		Http_code:	http_code,
		Res:		res,
	}, IDEM_EXPIRES); err != nil {
		panic(err)
	}
}

func Idempotency(r *http.Request, uid string) *idem_local {
	if !rdb.Connected() {
		panic("Redis is not connected")
	}
	
	l := &idem_local{}
	
	//	Check if header is provided
	key := r.Header.Get(IDEM_HEADER_KEY)
	if key == "" {
		l.error 		= fmt.Sprintf("%s header must be provided", IDEM_HEADER_KEY)
		l.http_code 	= http.StatusNotAcceptable
		return l
	}
	
	//	Check if value has the right length
	if len(key) > IDEM_HEADER_LENGTH {
		l.error 		= fmt.Sprintf("%s header value must not be longer than %d chars", IDEM_HEADER_KEY, IDEM_HEADER_LENGTH)
		l.http_code 	= http.StatusNotAcceptable
		return l
	}
	
	//	Check if request is duplicate
	var ref idem_remote
	l.hash = fmt.Sprintf(IDEM_HASH, uid, key)
	if rdb.Hgetall(r.Context(), l.hash, &ref) {
		l.http_code = ref.Http_code
		l.res 		= ref.Res
	}
	return l
}