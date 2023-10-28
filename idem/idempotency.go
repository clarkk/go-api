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
	IDEM_HEADER_KEY 	= "X-Idempotency-Key"
	IDEM_HEADER_CACHED 	= "X-Idempotency-Key-Cached"
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

//	Initiates idempotency
func Init(r *http.Request, uid string) *Idempotency {
	if !rdb.Connected() {
		panic("Redis is not connected")
	}
	
	d := &Idempotency{}
	
	//	Check if idempotency key header is provided
	key := r.Header.Get(IDEM_HEADER_KEY)
	if key == "" {
		d.http_code 	= http.StatusNotAcceptable
		d.error 		= errors.New(fmt.Sprintf("%s header must be provided", IDEM_HEADER_KEY))
		return d
	}
	
	//	Check if idempotency key value has the right length
	if len(key) > IDEM_HEADER_LENGTH {
		d.http_code 	= http.StatusNotAcceptable
		d.error 		= errors.New(fmt.Sprintf("%s header value must not be longer than %d chars", IDEM_HEADER_KEY, IDEM_HEADER_LENGTH))
		return d
	}
	
	//	Check if idempotency key is a duplicate and fetch response from cache
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

//	Returns error
func (d *Idempotency) Error() (int, error) {
	if d.error == nil {
		return 0, nil
	}
	return d.http_code, d.error
}

//	Get cached idempotency response
func (d *Idempotency) Cache() (int, string, head.Header){
	if d.http_code != 0 {
		return d.http_code, d.res, head.Header{
			IDEM_HEADER_CACHED,
			head.GMT_unix_time(d.time),
		}
	}
	return 0, "", head.Header{}
}

//	Cache response with idempotency key as JSON
func (d *Idempotency) Store_JSON(http_code int, res map[string]interface{}){
	if !store_http_codes(http_code){
		return
	}
	
	b, err := json.Marshal(res)
	if err != nil {
		panic("Idempotency store JSON encode: "+err.Error())
	}
	d.Store(http_code, string(b))
}

//	Cache response with idempotency key
func (d *Idempotency) Store(http_code int, res string){
	if !store_http_codes(http_code){
		return
	}
	
	if err := rdb.Hset(context.Background(), d.hash, cache{
		Time:		time.Now().Unix(),
		Http_code:	http_code,
		Res:		res,
	}, IDEM_EXPIRES); err != nil {
		panic("Could not store idempotency response: "+err.Error())
	}
}

func store_http_codes(http_code int) bool {
	return http_code == http.StatusOK
}