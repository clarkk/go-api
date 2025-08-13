package idem

import (
	"fmt"
	"time"
	"context"
	"net/http"
	"github.com/go-json-experiment/json"
	"github.com/clarkk/go-api"
	"github.com/clarkk/go-api/head"
	"github.com/clarkk/go-util/rdb"
)

const (
	HEADER_KEY 		= "Idempotency-Key"
	HEADER_TIME		= "Idempotency-Timestamp"
	HEADER_REPLAYED	= "Idempotency-Replayed"
	HEADER_EXPIRY 	= "Idempotency-Expiry"
	HEADER_LENGTH 	= 40
	EXPIRES 		= 60 * 60 * 24
	HASH 			= "API-IDEM:%s:%s:%s"
)

type (
	Idempotency struct {
		a				*api.Request
		required		bool
		key				string
		hash 			string
		time 			int64
		http_code 		int
		content_type	string
		res 			string
		req_hash		string
	}
	
	cache struct {
		Time 			int64 	`redis:"time"`
		Http_code 		int 	`redis:"http_code"`
		Content_type	string	`redis:"content_type"`
		Res 			string 	`redis:"res"`
		Req_hash		string	`redis:"req_hash"`
	}
)

//	New idempotency
func New(a *api.Request, uid string, required bool) (*Idempotency, error){
	if !rdb.Connected() {
		panic("Redis is not connected")
	}
	
	d := &Idempotency{
		a:			a,
		required:	required,
	}
	
	//	Check if idempotency key header is provided
	d.key = a.Request_header(HEADER_KEY)
	if d.key == "" {
		if !d.required {
			return d, nil
		}
		err := fmt.Errorf("%s header required", HEADER_KEY)
		a.Error(http.StatusNotAcceptable, err)
		return nil, err
	}
	
	//	Check if idempotency key value has the right length
	if len(d.key) > HEADER_LENGTH {
		err := fmt.Errorf("%s header value can not be longer than %d chars", HEADER_KEY, HEADER_LENGTH)
		a.Error(http.StatusNotAcceptable, err)
		return nil, err
	}
	
	//	Generate request hash
	var err error
	d.req_hash, err = a.Idempotency_hash()
	if err != nil {
		a.Error(http.StatusInternalServerError, nil)
		return nil, err
	}
	
	//	Check if idempotency key is a replay and fetch response from cache
	var res cache
	d.hash = fmt.Sprintf(HASH, uid, d.key, a.Request_URL_path())
	if err := rdb.Hgetall(a.Context(), d.hash, &res); err != nil {
		return nil, err
	}
	if res.Http_code != 0 {
		//	Check if idempotency key has been re-used in another request
		if d.req_hash != res.Req_hash {
			err := fmt.Errorf("Idempotency key has been re-used in another request")
			a.Error(http.StatusConflict, err)
			return nil, err
		}
		
		d.time 			= res.Time
		d.http_code 	= res.Http_code
		d.content_type	= res.Content_type
		d.res 			= res.Res
	}
	return d, nil
}

//	Disallow idempotency
func Disallow(a *api.Request) bool {
	//	Check if idempotency key header is provided
	if a.Request_header(HEADER_KEY) != "" {
		a.Error(http.StatusNotAcceptable, fmt.Errorf("%s header not allowed", HEADER_KEY))
		return false
	}
	return true
}

//	Get cached idempotency response
func (d *Idempotency) Cached() bool {
	if d.http_code == 0 {
		return false
	}
	d.headers(true, d.time)
	d.a.Response(d.http_code, d.content_type, d.res)
	return true
}

//	Cache response with idempotency key as JSON and send response
func (d *Idempotency) Response_JSON(code int, res any){
	b, err := json.Marshal(res)
	if err != nil {
		panic("Idempotency response JSON encode: "+err.Error())
	}
	s := string(b)
	if d.store_response(code){
		t := time_unix()
		go d.store_redis(code, head.TYPE_JSON, s, t)
		d.headers(false, t)
	}
	d.a.Response(code, head.TYPE_JSON, s)
}

//	Cache response with idempotency key and send response
func (d *Idempotency) Response(code int, content_type, res string){
	if d.store_response(code){
		t := time_unix()
		go d.store_redis(code, content_type, res, t)
		d.headers(false, t)
	}
	d.a.Response(code, content_type, res)
}

func (d *Idempotency) headers(replay bool, t int64){
	d.a.Header(HEADER_KEY, d.key)
	d.a.Header(HEADER_TIME, head.GMT_unix_time(t))
	if replay {
		d.a.Header(HEADER_REPLAYED, "true")
	} else {
		d.a.Header(HEADER_REPLAYED, "false")
	}
	d.a.Header(HEADER_EXPIRY, head.GMT_unix_time(t + EXPIRES))
}

func (d *Idempotency) store_redis(http_code int, content_type, res string, t int64){
	if err := rdb.Hset(context.Background(), d.hash, cache{
		Time:			t,
		Http_code:		http_code,
		Content_type:	content_type,
		Res:			res,
		Req_hash:		d.req_hash,
	}, EXPIRES); err != nil {
		panic("Could not store idempotency response: "+err.Error())
	}
}

func (d *Idempotency) store_response(http_code int) bool {
	if !d.required && d.key == "" {
		return false
	}
	return http_code == http.StatusOK
}

func time_unix() int64 {
	return time.Now().Unix()
}