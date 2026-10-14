package api

import (
	"fmt"
	"errors"
	"syscall"
	"context"
	"net/http"
	"encoding/json/v2"
	"encoding/json/jsontext"
	"github.com/clarkk/go-api/errin"
	"github.com/clarkk/go-api/head"
	"github.com/clarkk/go-api/map_json"
)

type (
	Response_result struct {
		Result		any 					`json:"result"`
		Limit		*Limit					`json:"limit,omitzero"`
	}
	
	response_error struct {
		Error 		*map_json.Map			`json:"error,omitzero"`
		Warning 	*map_json.Map			`json:"warning,omitzero"`
		Limit		*Limit					`json:"limit,omitzero"`
	}
	
	response_bulk_errors struct {
		Errors 			[]*map_json.Map		`json:"errors,omitzero"`
		Semantic_errors []*string			`json:"semantic_errors,omitzero"`
	}
)

//	Set header
func (a *Request) Header(key, value string){
	if a.w.Sent_header() {
		panic("Header already sent. Can not set header: "+key)
	}
	a.w.Header().Set(key, value)
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
	a.Error_limit(status, err, nil)
}

//	Error JSON response with limit
func (a *Request) Error_limit(status int, err error, limit *Limit){
	if a.w.Sent_header() {
		panic("HTTP header already sent")
	}
	if err == nil {
		err = errors.New(http.StatusText(status))
	}
	errs := map_json.New()
	errs.Set("request", err.Error())
	a.Response_JSON(status, response_error{
		Error:	errs,
		Limit:	limit,
	})
}

//	Errors JSON response
func (a *Request) Errors(status int, errs errin.Map){
	if a.w.Sent_header() {
		panic("HTTP header already sent")
	}
	a.Response_JSON(status, response_error{
		Error: errs.Output(),
	})
}

//	Warnings JSON response
func (a *Request) Warnings(status int, errs errin.Map){
	if a.w.Sent_header() {
		panic("HTTP header already sent")
	}
	a.Response_JSON(status, response_error{
		Warning: errs.Output(),
	})
}

//	Errors JSON response
func (a *Request) Bulk_errors(status int, bulk_errs []errin.Map){
	if a.w.Sent_header() {
		panic("HTTP header already sent")
	}
	bulk := make([]*map_json.Map, len(bulk_errs))
	for i, errs := range bulk_errs {
		bulk[i] = errs.Output()
	}
	a.Response_JSON(status, response_bulk_errors{
		Errors: bulk,
	})
}

//	Errors JSON response
func (a *Request) Bulk_semantic_errors(status int, bulk_errs []error){
	if a.w.Sent_header() {
		panic("HTTP header already sent")
	}
	bulk := make([]*string, len(bulk_errs))
	for i, err := range bulk_errs {
		if err != nil {
			s := err.Error()
			bulk[i] = &s
		}
	}
	a.Response_JSON(status, response_bulk_errors{
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
	a.w.WriteHeader(status)
}

//	Write JSON response
func (a *Request) write_JSON(res any){
	if err := json.MarshalWrite(a.w, res); err != nil {
		var (
			semanticErr *json.SemanticError
			syntaxErr *jsontext.SyntacticError
		)
		if errors.As(err, &semanticErr) || errors.As(err, &syntaxErr) {
			panic(fmt.Errorf("API response JSON marshal: %w", err))
		}
		
		if disconnect_error(err) {
			panic(http.ErrAbortHandler)
		}
		
		panic(fmt.Errorf("API response JSON write: %w", err))
	}
	a.exec_deferred()
}

//	Write response
func (a *Request) write(res string){
	if _, err := a.w.Write([]byte(res)); err != nil {
		if disconnect_error(err) {
			panic(http.ErrAbortHandler)
		}
		
		panic(fmt.Errorf("API response write: %w", err))
	}
	a.exec_deferred()
}

func (a *Request) exec_deferred(){
	if a.deferred != nil {
		a.deferred(a)
	}
}

func disconnect_error(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, syscall.EPIPE) || errors.Is(err, syscall.ECONNRESET)
}