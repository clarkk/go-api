package idem

import (
	"fmt"
	"time"
	"context"
	"net/http"
	"github.com/go-errors/errors"
	"github.com/go-json-experiment/json"
	"github.com/clarkk/go-api/head"
	"github.com/clarkk/go-util/rdb"
)

const (
	IDEM_HEADER_KEY 	= "Idempotency-Key"
	IDEM_HEADER_CACHED 	= "Idempotency-Key-Cached"
	IDEM_HEADER_LENGTH 	= 40
	IDEM_EXPIRES 		= 60 * 60 * 24
	IDEM_HASH 			= "API-IDEM:%s:%s"
)

type (
	Idempotency struct {
		error 		error
		hash 		string
		time 		int64
		http_code 	int
		res 		string
	}
	
	cache struct {
		Time 		int64 	`redis:"time"`
		Http_code 	int 	`redis:"http_code"`
		Res 		string 	`redis:"res"`
	}
)

func (l *Idempotency) Cache() (int, string, head.Header){
	if l.http_code != 0 {
		return l.http_code, l.res, head.Header{
			IDEM_HEADER_CACHED,
			head.GMT_unix_time(l.time),
		}
	}
	return 0, "", head.Header{}
}

func (l *Idempotency) Error() (int, error) {
	if l.error == nil {
		return 0, nil
	}
	return l.http_code, l.error
}

func (l *Idempotency) Store_JSON(http_code int, res map[string]interface{}){
	if !store_http_codes(http_code){
		return
	}
	
	b, err := json.Marshal(res)
	if err != nil {
		panic("Idempotency store JSON encode: "+err.Error())
	}
	l.Store(http_code, string(b))
}

func (l *Idempotency) Store(http_code int, res string){
	if !store_http_codes(http_code){
		return
	}
	
	if err := rdb.Hset(context.Background(), l.hash, cache{
		Time:		time.Now().Unix(),
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
	
	d := &Idempotency{}
	
	//	Check if header is provided
	key := r.Header.Get(IDEM_HEADER_KEY)
	if key == "" {
		d.http_code 	= http.StatusNotAcceptable
		d.error 		= errors.New(fmt.Sprintf("%s header must be provided", IDEM_HEADER_KEY))
		return d
	}
	
	//	Check if value has the right length
	if len(key) > IDEM_HEADER_LENGTH {
		d.http_code 	= http.StatusNotAcceptable
		d.error 		= errors.New(fmt.Sprintf("%s header value must not be longer than %d chars", IDEM_HEADER_KEY, IDEM_HEADER_LENGTH))
		return d
	}
	
	//	Check if request is duplicate
	var ref cache
	d.hash = fmt.Sprintf(IDEM_HASH, uid, key)
	rdb.Hgetall(r.Context(), d.hash, &ref)
	if ref.Http_code != 0 {
		d.time 		= ref.Time
		d.http_code = ref.Http_code
		d.res 		= ref.Res
	}
	return d
}

func store_http_codes(http_code int) bool {
	return http_code == http.StatusOK
}