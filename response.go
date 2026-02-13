package api

import (
	"fmt"
	"sync"
	"net/http"
	"compress/gzip"
	"github.com/go-json-experiment/json"
	"github.com/clarkk/go-api/errin"
	"github.com/clarkk/go-api/head"
	"github.com/clarkk/go-api/map_json"
)

var gzip_pool = sync.Pool{
	New: func() any {
		return gzip.NewWriter(nil)
	},
}

type (
	Response_result struct {
		Result		any 				`json:"result"`
		Limit		*Limit				`json:"limit,omitempty"`
	}
	
	response_error struct {
		Error 		*errin.Map			`json:"error,omitempty"`
		Warning 	*errin.Map			`json:"warning,omitempty"`
	}
	
	response_bulk_errors struct {
		Errors 			[]*errin.Map	`json:"errors,omitempty"`
		Semantic_errors []*string		`json:"semantic_errors,omitempty"`
	}
)

//	Set header
func (a *Request) Header(key, value string){
	if a.w.Sent_header() {
		panic("Header already sent. Can not set header: "+key)
	}
	a.header[key] = value
}

//	Send JSON response (encode output)
func (a *Request) Response_JSON(status int, res any){
	a.Header(head.CONTENT_TYPE, head.TYPE_JSON)
	a.write_header(status)
	a.write_JSON(res)
}

//	Send response
func (a *Request) Response(status int, content_type, res string){
	a.Header(head.CONTENT_TYPE, content_type)
	a.write_header(status)
	a.write(res)
}

//	Error JSON response
func (a *Request) Errorf(status int, s string, args... any){
	a.Error(status, fmt.Errorf(s, args...))
}

//	Error JSON response
func (a *Request) Error(status int, err error){
	if a.w.Sent_header() {
		//	TODO: handle panics/errors AFTER headers are sent
		panic("HTTP header already sent")
	}
	a.Header(head.CONTENT_TYPE, head.TYPE_JSON)
	a.write_header(status)
	if err == nil {
		err = fmt.Errorf(http.StatusText(status))
	}
	errs := map_json.New()
	errs.Set("request", err.Error())
	a.write_JSON(response_error{
		Error: errs,
	})
}

//	Errors JSON response
func (a *Request) Errors(status int, errs *errin.Map){
	if a.w.Sent_header() {
		//	TODO: handle panics/errors AFTER headers are sent
		panic("HTTP header already sent")
	}
	a.Header(head.CONTENT_TYPE, head.TYPE_JSON)
	a.write_header(status)
	a.write_JSON(response_error{
		Error: errs,
	})
}

//	Warnings JSON response
func (a *Request) Warnings(status int, errs *errin.Map){
	if a.w.Sent_header() {
		//	TODO: handle panics/errors AFTER headers are sent
		panic("HTTP header already sent")
	}
	a.Header(head.CONTENT_TYPE, head.TYPE_JSON)
	a.write_header(status)
	a.write_JSON(response_error{
		Warning: errs,
	})
}

//	Errors JSON response
func (a *Request) Bulk_errors(status int, bulk_errs []*errin.Map){
	if a.w.Sent_header() {
		//	TODO: handle panics/errors AFTER headers are sent
		panic("HTTP header already sent")
	}
	a.Header(head.CONTENT_TYPE, head.TYPE_JSON)
	a.write_header(status)
	a.write_JSON(response_bulk_errors{
		Errors: bulk_errs,
	})
}

//	Errors JSON response
func (a *Request) Bulk_semantic_errors(status int, bulk_errs []error){
	if a.w.Sent_header() {
		//	TODO: handle panics/errors AFTER headers are sent
		panic("HTTP header already sent")
	}
	a.Header(head.CONTENT_TYPE, head.TYPE_JSON)
	a.write_header(status)
	bulk := make([]*string, len(bulk_errs))
	for i, err := range bulk_errs {
		if err != nil {
			s := err.Error()
			bulk[i] = &s
		}
	}
	a.write_JSON(response_bulk_errors{
		Semantic_errors: bulk,
	})
}

func (a *Request) Status() int {
	return a.w.Status()
}

func (a *Request) Sent() int {
	return a.w.Sent()
}

//	Send header
func (a *Request) write_header(status int){
	if a.accept_gzip {
		a.Header(head.CONTENT_ENCODING, head.ENCODING_GZIP)
	}
	header := a.w.Header()
	for key, value := range a.header {
		header.Set(key, value)
	}
	a.w.WriteHeader(status)
}

//	Write JSON response
func (a *Request) write_JSON(res any){
	if a.accept_gzip {
		gz := gzip_pool.Get().(*gzip.Writer)
		gz.Reset(a.w)
		defer gzip_pool.Put(gz)
		defer gz.Close()
		
		if err := json.MarshalWrite(gz, res); err != nil {
			switch t := err.(type) {
			case *json.SemanticError:
				panic("API response JSON marshal (gzip): "+t.Error())
			}
		}
	} else {
		if err := json.MarshalWrite(a.w, res); err != nil {
			switch t := err.(type) {
			case *json.SemanticError:
				panic("API response JSON marshal: "+t.Error())
			}
		}
	}
	if a.deferred != nil {
		a.deferred(a)
	}
}

//	Write response
func (a *Request) write(res string){
	if a.accept_gzip {
		gz := gzip_pool.Get().(*gzip.Writer)
		gz.Reset(a.w)
		defer gzip_pool.Put(gz)
		defer gz.Close()
		
		gz.Write([]byte(res))
	} else {
		a.w.Write([]byte(res))
	}
	if a.deferred != nil {
		a.deferred(a)
	}
}