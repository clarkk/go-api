package api

import (
	"strings"
	"net/http"
	"compress/gzip"
	"github.com/go-errors/errors"
	"github.com/go-json-experiment/json"
	"github.com/clarkk/go-api/idem"
	"github.com/clarkk/go-util/serv"
)

const (
	CONTENT_TYPE 		= "Content-Type"
	CONTENT_ENCODING 	= "Content-Encoding"
	
	TYPE_JSON 			= "application/json"
	TYPE_FORM_DATA 		= "application/x-www-form-urlencoded"
	
	ACCEPT_ENCODING 	= "Accept-Encoding"
	ENCODING_GZIP 		= "gzip"
)

type (
	list 	map[string]string
	body 	map[string]interface{}
	
	request struct {
		w 				http.ResponseWriter
		r 				*http.Request
		accept_gzip 	bool
		body 			body
		idem 			*idem.Idempotency
	}
	
	response_error struct {
		Error 	list 			`json:"error"`
	}
	
	/*response_result struct {
		Result 	[]interface{} 	`json:"result"`
	}*/
)

func NewRequest(w http.ResponseWriter, r *http.Request) *request{
	w.Header().Set(CONTENT_TYPE, TYPE_JSON)
	return &request{
		w,
		r,
		accept_gzip(r),
		body{},
		nil,
	}
}

func (a *request) Idempotency(uid string) *idem.Idempotency {
	a.idem = idem.Init(a.r, uid)
	return a.idem
}

func (a *request) Body() body {
	return a.body
}

func (a *request) Parse_body(post_limit int64) (int, error){
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

func (a *request) Error(code int, error error){
	b, err := json.Marshal(response_error{
		Error: list{"request": error.Error()},
	})
	if err != nil {
		panic("API error response JSON encode: "+err.Error())
	}
	
	a.w.WriteHeader(code)
	a.w.Write(b)
}

func (a *request) Response_JSON(code int, res body){
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

func (a *request) Response(code int, res string){
	a.write_header(code)
	if a.accept_gzip {
		gz := gzip.NewWriter(a.w)
		defer gz.Close()
		gz.Write([]byte(res))
	}else{
		a.w.Write([]byte(res))
	}
}

func (a *request) write_header(code int){
	if a.accept_gzip {
		a.w.Header().Set(CONTENT_ENCODING, ENCODING_GZIP)
	}
	a.w.WriteHeader(code)
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