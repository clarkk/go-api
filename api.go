package api

import (
	"fmt"
	"log"
	"strings"
	"context"
	"net/http"
	"runtime/debug"
	"compress/gzip"
	"github.com/go-errors/errors"
	"github.com/go-json-experiment/json"
	"github.com/clarkk/go-util/serv"
)

const (
	CONTENT_TYPE 		= "Content-Type"
	CONTENT_ENCODING 	= "Content-Encoding"
	
	TYPE_JSON 			= "application/json"
	TYPE_FORM_DATA 		= "application/x-www-form-urlencoded"
	
	ACCEPT_ENCODING 	= "Accept-Encoding"
	ENCODING_GZIP 		= "gzip"
	
	CTX_API 	ctx_key = ""
)

type (
	Request struct {
		w 				http.ResponseWriter
		r 				*http.Request
		accept_gzip 	bool
		body 			body
		
		header_sent 	bool
		header 			list
	}
	
	list 	map[string]string
	body 	map[string]interface{}
	
	response_error struct {
		Error 	list 	`json:"error"`
	}
	
	ctx_key 			string
	
	/*response_result struct {
		Result 	[]interface{} 	`json:"result"`
	}*/
)

func NewRequest(w http.ResponseWriter, r *http.Request) *Request{
	return &Request{
		w:				w,
		r:				r,
		accept_gzip:	accept_gzip(r),
		body:			body{},
		
		header:			list{},
	}
}

func (a *Request) Recover(){
	if r := recover(); r != nil {
		a.Error(http.StatusBadRequest, errors.New("Unexpected error"))
		log.Println(r, "\n"+string(debug.Stack()))
	}
}

func (a *Request) Wrap_ctx() *http.Request {
	return a.r.WithContext(context.WithValue(a.r.Context(), CTX_API, a))
}

func Wrap_get_api(r *http.Request) *Request {
	return r.Context().Value(CTX_API).(*Request)
}

func (a *Request) Auth() (code int, error error){
	key := a.r.Header.Get("X-Key")
	hash := a.r.Header.Get("X-Hash")
	fmt.Println("key:", key)
	fmt.Println("hash:", hash)
	
	return 0, nil
}

func (a *Request) Body() body {
	return a.body
}

func (a *Request) Parse_body(post_limit int64) (int, error){
	body_bytes, err := serv.Post_limit(a.w, a.r, post_limit)
	if err != nil {
		return http.StatusRequestEntityTooLarge, errors.New(http.StatusText(http.StatusRequestEntityTooLarge))
	}
	
	switch a.r.Header.Get(CONTENT_TYPE) {
	case TYPE_JSON:
		if json.Unmarshal(body_bytes, &a.body) != nil {
			return http.StatusBadRequest, errors.New(http.StatusText(http.StatusBadRequest))
		}
		
	/*case TYPE_FORM_DATA:
		if a.r.ParseForm() != nil {
			return http.StatusBadRequest, errors.New(http.StatusText(http.StatusBadRequest))
		}*/
		
	default:
		return http.StatusUnsupportedMediaType, errors.New("POST format not supported")
	}
	
	return 0, nil
}

func (a *Request) Error(code int, error error){
	b, err := json.Marshal(response_error{
		Error: list{"request": error.Error()},
	})
	if err != nil {
		panic("API error response JSON encode: "+err.Error())
	}
	
	a.w.WriteHeader(code)
	//a.write_header(code)
	a.w.Write(b)
}

func (a *Request) Header(key string, value string){
	if a.header_sent {
		panic("Header already sent. Can not set header: "+key)
	}
	a.header[key] = value
}

func (a *Request) Response_JSON(code int, res body){
	a.write_header(code)
	
	if a.accept_gzip {
		gz := gzip.NewWriter(a.w)
		defer gz.Close()
		if err := json.MarshalWrite(gz, res); err != nil {
			panic("API response JSON encode (gzip): "+err.Error())
		}
	}else{
		if err := json.MarshalWrite(a.w, res); err != nil {
			panic("API response JSON encode: "+err.Error())
		}
	}
}

func (a *Request) Response(code int, res string){
	a.write_header(code)
	
	if a.accept_gzip {
		gz := gzip.NewWriter(a.w)
		defer gz.Close()
		gz.Write([]byte(res))
	}else{
		a.w.Write([]byte(res))
	}
}

func (a *Request) write_header(code int){
	a.Header(CONTENT_TYPE, TYPE_JSON)
	if a.accept_gzip {
		a.Header(CONTENT_ENCODING, ENCODING_GZIP)
	}
	
	header := a.w.Header()
	for key, value := range a.header {
		header.Set(key, value)
	}
	
	a.w.WriteHeader(code)
	a.header_sent = true
}

func accept_gzip(r *http.Request) bool {
	header := r.Header.Get(ACCEPT_ENCODING)
	if header == "" {
		return false
	}
	for _, value := range strings.Split(header, ",") {
		if strings.TrimSpace(value) == ENCODING_GZIP {
			return true
		}
	}
	return false
}