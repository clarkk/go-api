package api

import (
	"fmt"
	"log"
	"context"
	"strings"
	"net/http"
	"path/filepath"
	"compress/gzip"
	"github.com/go-errors/errors"
	"github.com/go-json-experiment/json"
	"github.com/clarkk/go-api/head"
	"github.com/clarkk/go-util/serv/req"
)

type (
	Request struct {
		w 				http.ResponseWriter
		r 				*http.Request
		
		handle_gzip		bool
		accept_gzip 	bool
		
		header_sent 	bool
		header 			List
	}
	
	Response_result struct {
		Result			any 		`json:"result"`
	}
	
	List map[string]string
	
	response_error struct {
		Error 			List		`json:"error"`
	}
	
	response_bulk_errors struct {
		Errors 			[]*List		`json:"errors,omitempty"`
		Semantic_errors []*string	`json:"semantic_errors,omitempty"`
	}
)

func New(w http.ResponseWriter, r *http.Request, handle_gzip bool) *Request {
	return &Request{
		w:				w,
		r:				r,
		handle_gzip:	handle_gzip,
		accept_gzip:	accept_gzip(r, handle_gzip),
		header:			List{},
	}
}

//	Recover from panic inside route handler
func (a *Request) Recover(){
	if err := recover(); err != nil {
		a.Errorf(http.StatusInternalServerError, "Unexpected error")
		log.Println(errors.Wrap(err, 2).ErrorStack())
	}
}

//	Get request context
func (a *Request) Request_context() context.Context {
	return a.r.Context()
}

//	Get request header
func (a *Request) Request_header(name string) string {
	return a.r.Header.Get(name)
}

//	Get request URL path
func (a *Request) Request_URL_path() string {
	return filepath.Clean(a.r.URL.Path)
}

//	Parse request POST body
func (a *Request) Request(post_limit int) ([]byte, error){
	b, err := req.Post_limit_read(a.w, a.r, post_limit)
	if err != nil {
		if error_request_too_large(err) {
			a.Error(http.StatusRequestEntityTooLarge, nil)
			return b, err
		}
		a.Errorf(http.StatusInternalServerError, "Unable to read request body")
		return b, err
	}
	return b, nil
}

//	Parse request POST body as JSON
func (a *Request) Request_JSON(post_limit int, input any) error {
	b, err := a.request_JSON(post_limit)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, input, json.RejectUnknownMembers(true)); err != nil {
		var serr *json.SemanticError
		if errors.As(err, &serr) {
			err = parse_unmarshal_json_error(serr, b, input)
			a.Error(http.StatusBadRequest, err)
			return err
		}
		a.Errorf(http.StatusBadRequest, "Unable to unmarshal JSON")
		return err
	}
	return nil
}

//	Parse request POST body as JSON array
func (a *Request) Request_JSON_slice(post_limit int, input any) (error, []error, bool){
	b, err := a.request_JSON(post_limit)
	if err != nil {
		return err, nil, false
	}
	if err := json.Unmarshal(b, input, json.RejectUnknownMembers(true)); err != nil {
		var serr *json.SemanticError
		if errors.As(err, &serr) {
			err, slice_errs := parse_unmarshal_json_slice_error(serr, b, input)
			if err != nil {
				a.Error(http.StatusBadRequest, err)
			} else {
				a.Bulk_semantic_errors(http.StatusBadRequest, slice_errs)
			}
			return err, slice_errs, false
		}
		a.Errorf(http.StatusBadRequest, "Unable to unmarshal JSON")
		return err, nil, false
	}
	return nil, nil, true
}

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

func (a *Request) request_JSON(post_limit int) ([]byte, error){
	b, err := req.Post_limit_read(a.w, a.r, post_limit)
	if err != nil {
		if error_request_too_large(err) {
			a.Error(http.StatusRequestEntityTooLarge, nil)
			return b, err
		}
		a.Errorf(http.StatusInternalServerError, "Unable to read request body")
		return b, err
	}
	if !head.Request_JSON(a.r) {
		a.Error(http.StatusUnsupportedMediaType, nil)
		return b, fmt.Errorf("Unsupported media type")
	}
	return b, nil
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

func accept_gzip(r *http.Request, handle_gzip bool) bool {
	if !handle_gzip {
		return false
	}
	
	header := r.Header.Get(head.ACCEPT_ENCODING)
	if header == "" {
		return false
	}
	for _, value := range strings.Split(header, ",") {
		if strings.TrimSpace(value) == head.ENCODING_GZIP {
			return true
		}
	}
	return false
}

func error_request_too_large(err error) bool {
	return err.Error() == "http: request body too large"
}