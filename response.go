package api

import (
	"fmt"
	"net/http"
	"compress/gzip"
	"github.com/go-errors/errors"
	"github.com/go-json-experiment/json"
	"github.com/clarkk/go-api/head"
)

type (
	Response_result struct {
		Result			any 		`json:"result"`
	}
	
	response_error struct {
		Error 			List		`json:"error"`
	}
	
	response_bulk_errors struct {
		Errors 			[]*List		`json:"errors,omitempty"`
		Semantic_errors []*string	`json:"semantic_errors,omitempty"`
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
	}
	list := List{}
	for key, err := range errs {
		list[key] = err.Error()
	}
	a.write_JSON(response_error{
		Error: list,
	})
}

//	Errors JSON response
func (a *Request) Bulk_errors(code int, bulk_errors []map[string]error){
	if !a.header_sent {
		a.Header(head.CONTENT_TYPE, head.TYPE_JSON)
		a.write_header(code)
	} else {
		//	TODO: handle panics/errors AFTER headers are sent
	}
	bulk := make([]*List, len(bulk_errors))
	for i, errs := range bulk_errors {
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
func (a *Request) Bulk_semantic_errors(code int, bulk_errors []error){
	if !a.header_sent {
		a.Header(head.CONTENT_TYPE, head.TYPE_JSON)
		a.write_header(code)
	} else {
		//	TODO: handle panics/errors AFTER headers are sent
	}
	bulk := make([]*string, len(bulk_errors))
	for i, err := range bulk_errors {
		if err != nil {
			s := err.Error()
			bulk[i] = &s
		}
	}
	a.write_JSON(response_bulk_errors{
		Semantic_errors: bulk,
	})
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
	a.header_sent = true
}

//	Write JSON response
func (a *Request) write_JSON(res any){
	if a.accept_gzip {
		gz := gzip.NewWriter(a.w)
		defer gz.Close()
		if err := json.MarshalWrite(gz, res); err != nil {
			var serr *json.SemanticError
			if errors.As(err, &serr) {
				panic("API response JSON encode (gzip): "+err.Error())
			}
		}
	} else {
		if err := json.MarshalWrite(a.w, res); err != nil {
			var serr *json.SemanticError
			if errors.As(err, &serr) {
				panic("API response JSON encode: "+err.Error())
			}
		}
	}
}

//	Write response
func (a *Request) write(res string){
	if a.accept_gzip {
		gz := gzip.NewWriter(a.w)
		defer gz.Close()
		gz.Write([]byte(res))
	} else {
		a.w.Write([]byte(res))
	}
}