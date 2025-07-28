package api

import (
	"fmt"
	"net/http"
	"compress/gzip"
	"github.com/go-json-experiment/json"
	"github.com/clarkk/go-api/head"
)

type (
	Response_result struct {
		Result		any 		`json:"result"`
		Limit		*Limit		`json:"limit,omitempty"`
	}
	
	response_error struct {
		Error 		List		`json:"error,omitempty"`
		Warning 	List		`json:"warning,omitempty"`
	}
	
	response_bulk_errors struct {
		Errors 			[]*List		`json:"errors,omitempty"`
		Semantic_errors []*string	`json:"semantic_errors,omitempty"`
	}
	
	response_writer struct {
		http.ResponseWriter
		bytes_sent 	int
	}
)

//	Set header
func (a *Request) Header(key, value string){
	if a.header_sent {
		panic("Header already sent. Can not set header: "+key)
	}
	a.header[key] = value
}

//	Send JSON response (encode output)
func (a *Request) Response_JSON(code int, res any){
	a.Header(head.CONTENT_TYPE, head.TYPE_JSON)
	a.write_header(code)
	a.write_JSON(res)
}

//	Send response
func (a *Request) Response(code int, content_type, res string){
	a.Header(head.CONTENT_TYPE, content_type)
	a.write_header(code)
	a.write(res)
}

//	Error JSON response
func (a *Request) Errorf(code int, s string, args... any){
	a.Error(code, fmt.Errorf(s, args...))
}

//	Error JSON response
func (a *Request) Error(code int, err error){
	if err == nil {
		err = fmt.Errorf(http.StatusText(code))
	}
	if !a.header_sent {
		a.Header(head.CONTENT_TYPE, head.TYPE_JSON)
		a.write_header(code)
	} else {
		//	TODO: handle panics/errors AFTER headers are sent
		panic("HTTP header already sent")
	}
	a.write_JSON(response_error{
		Error: List{"request": err.Error()},
	})
}

//	Errors JSON response
func (a *Request) Errors(code int, errs map[string]error){
	if !a.header_sent {
		a.Header(head.CONTENT_TYPE, head.TYPE_JSON)
		a.write_header(code)
	} else {
		//	TODO: handle panics/errors AFTER headers are sent
		panic("HTTP header already sent")
	}
	list := List{}
	for key, err := range errs {
		list[key] = err.Error()
	}
	a.write_JSON(response_error{
		Error: list,
	})
}

//	Warnings JSON response
func (a *Request) Warnings(code int, errs map[string]error){
	if !a.header_sent {
		a.Header(head.CONTENT_TYPE, head.TYPE_JSON)
		a.write_header(code)
	} else {
		//	TODO: handle panics/errors AFTER headers are sent
		panic("HTTP header already sent")
	}
	list := List{}
	for key, err := range errs {
		list[key] = err.Error()
	}
	a.write_JSON(response_error{
		Warning: list,
	})
}

//	Errors JSON response
func (a *Request) Bulk_errors(code int, bulk_errs []map[string]error){
	if !a.header_sent {
		a.Header(head.CONTENT_TYPE, head.TYPE_JSON)
		a.write_header(code)
	} else {
		//	TODO: handle panics/errors AFTER headers are sent
		panic("HTTP header already sent")
	}
	bulk := make([]*List, len(bulk_errs))
	for i, errs := range bulk_errs {
		if errs != nil {
			l := List{}
			for key, err := range errs {
				l[key] = err.Error()
			}
			bulk[i] = &l
		}
	}
	a.write_JSON(response_bulk_errors{
		Errors: bulk,
	})
}

//	Errors JSON response
func (a *Request) Bulk_semantic_errors(code int, bulk_errs []error){
	if !a.header_sent {
		a.Header(head.CONTENT_TYPE, head.TYPE_JSON)
		a.write_header(code)
	} else {
		//	TODO: handle panics/errors AFTER headers are sent
		panic("HTTP header already sent")
	}
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

func (a *Request) Code() int {
	return a.code
}

func (a *Request) Sent() int {
	return a.bytes_sent
}

//	Send header
func (a *Request) write_header(code int){
	if a.accept_gzip {
		a.Header(head.CONTENT_ENCODING, head.ENCODING_GZIP)
	}
	header := a.w.Header()
	for key, value := range a.header {
		header.Set(key, value)
	}
	a.w.WriteHeader(code)
	a.code 			= code
	a.header_sent 	= true
}

//	Write JSON response
func (a *Request) write_JSON(res any){
	//	Wrap writer to count bytes sent
	w := &response_writer{
		ResponseWriter: a.w,
	}
	if a.accept_gzip {
		gz := gzip.NewWriter(w)
		defer gz.Close()
		if err := json.MarshalWrite(gz, res); err != nil {
			switch t := err.(type) {
			case *json.SemanticError:
				panic("API response JSON marshal (gzip): "+t.Error())
			}
		}
	} else {
		if err := json.MarshalWrite(w, res); err != nil {
			switch t := err.(type) {
			case *json.SemanticError:
				panic("API response JSON marshal: "+t.Error())
			}
		}
	}
	a.bytes_sent = w.bytes_sent
	if a.deferred != nil {
		a.deferred(a)
	}
}

//	Write response
func (a *Request) write(res string){
	//	Wrap writer to count bytes sent
	w := &response_writer{
		ResponseWriter: a.w,
	}
	if a.accept_gzip {
		gz := gzip.NewWriter(w)
		defer gz.Close()
		gz.Write([]byte(res))
	} else {
		w.Write([]byte(res))
	}
	a.bytes_sent = w.bytes_sent
	if a.deferred != nil {
		a.deferred(a)
	}
}

func (r *response_writer) Write(b []byte) (int, error){
	n, err := r.ResponseWriter.Write(b)
	r.bytes_sent += n
	return n, err
}