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
		status		int
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
	if err == nil {
		err = fmt.Errorf(http.StatusText(status))
	}
	if !a.header_sent {
		a.Header(head.CONTENT_TYPE, head.TYPE_JSON)
		a.write_header(status)
	} else {
		//	TODO: handle panics/errors AFTER headers are sent
		panic("HTTP header already sent")
	}
	a.write_JSON(response_error{
		Error: List{"request": err.Error()},
	})
}

//	Errors JSON response
func (a *Request) Errors(status int, errs map[string]error){
	if !a.header_sent {
		a.Header(head.CONTENT_TYPE, head.TYPE_JSON)
		a.write_header(status)
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
func (a *Request) Warnings(status int, errs map[string]error){
	if !a.header_sent {
		a.Header(head.CONTENT_TYPE, head.TYPE_JSON)
		a.write_header(status)
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
func (a *Request) Bulk_errors(status int, bulk_errs []map[string]error){
	if !a.header_sent {
		a.Header(head.CONTENT_TYPE, head.TYPE_JSON)
		a.write_header(status)
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
func (a *Request) Bulk_semantic_errors(status int, bulk_errs []error){
	if !a.header_sent {
		a.Header(head.CONTENT_TYPE, head.TYPE_JSON)
		a.write_header(status)
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

func (a *Request) Status() int {
	return a.w.status
}

func (a *Request) Sent() int {
	return a.w.bytes_sent
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
	a.header_sent 	= true
}

//	Write JSON response
func (a *Request) write_JSON(res any){
	//	Wrap writer to count bytes sent
	if a.accept_gzip {
		gz := gzip.NewWriter(a.w)
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
	//	Wrap writer to count bytes sent
	w := &response_writer{
		ResponseWriter: a.w,
	}
	if a.accept_gzip {
		gz := gzip.NewWriter(a.w)
		defer gz.Close()
		gz.Write([]byte(res))
	} else {
		a.w.Write([]byte(res))
	}
	a.status		= w.status
	a.bytes_sent	= w.bytes_sent
	if a.deferred != nil {
		a.deferred(a)
	}
}

func (r *response_writer) WriteHeader(status int){
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *response_writer) Write(b []byte) (int, error){
	n, err := r.ResponseWriter.Write(b)
	r.bytes_sent += n
	return n, err
}