package api

import (
	"fmt"
	"path"
	"context"
	"strings"
	"net/http"
	"github.com/go-json-experiment/json"
	"github.com/clarkk/go-api/head"
	"github.com/clarkk/go-api/invalid_json"
	"github.com/clarkk/go-util/env"
	"github.com/clarkk/go-util/hash"
	"github.com/clarkk/go-util/serv"
	"github.com/clarkk/go-util/serv/req"
)

type (
	Request struct {
		w 				*serv.Writer
		r 				*http.Request
		e				*env.Environment
		
		handle_gzip		bool
		accept_gzip 	bool
		
		body_received	[]byte
		
		header 			List
		
		deferred 		func(*Request)
	}
	
	Input interface {
		Required() error
	}
	
	List map[string]string
)

func Input_required_error(s []string) error {
	if len(s) == 0 {
		return nil
	}
	return fmt.Errorf("Required fields: %s", strings.Join(s, ", "))
}

func New(w http.ResponseWriter, r *http.Request, handle_gzip bool) *Request {
	sw, ok := w.(*serv.Writer)
	if !ok {
		sw = serv.NewWriter(w)
	}
	
	return &Request{
		w:				sw,
		r:				r,
		e:				env.Request(r),
		handle_gzip:	handle_gzip,
		accept_gzip:	accept_gzip(r, handle_gzip),
		header:			List{},
	}
}

//	Recover from panic inside route handler
func (a *Request) Recover(){
	if r := recover(); r != nil {
		var err error
		switch t := r.(type) {
		case error:
			err = t
		default:
			err = fmt.Errorf("panic: %v", t)
		}
		
		code := http.StatusInternalServerError
		if !a.w.Sent_header() {
			a.Error(code, nil)
		}
		a.log(code, err, a.e)
	}
}

//	Set deferred function
func (a *Request) Defer(fn func(*Request)){
	a.deferred = fn
}

//	Get request method
func (a *Request) Method() string {
	return a.r.Method
}

//	Get request context
func (a *Request) Context() context.Context {
	return a.r.Context()
}

//	Get request header
func (a *Request) Request_header(name string) string {
	return a.r.Header.Get(name)
}

//	Get request URL path
func (a *Request) Request_URL_path() string {
	return path.Clean(a.r.URL.Path)
}

//	Parse request POST body
func (a *Request) Request(post_limit int) ([]byte, int, error){
	var err error
	a.body_received, err = req.Post_limit_read(a.w, a.r, post_limit)
	if err != nil {
		if error_request_too_large(err) {
			return nil, http.StatusRequestEntityTooLarge, fmt.Errorf("POST payload too large")
		}
		return nil, http.StatusInternalServerError, fmt.Errorf("Unable to read request body")
	}
	return a.body_received, 0, nil
}

//	Parse request POST body as JSON
func (a *Request) Request_JSON(post_limit int, input any) (int, error){
	b, code, err := a.request_JSON(post_limit)
	if err != nil {
		return code, err
	}
	if err := json.Unmarshal(b, input, json.RejectUnknownMembers(true)); err != nil {
		switch t := err.(type) {
		case *json.SemanticError:
			return http.StatusBadRequest, invalid_json.Fields(t, b, input)
		}
		return http.StatusBadRequest, fmt.Errorf("Unable to unmarshal JSON")
	}
	return 0, nil
}

//	Parse request POST body as JSON and require input
func (a *Request) Request_JSON_required(post_limit int, input Input) (int, error){
	if code, err := a.Request_JSON(post_limit, input); code != 0 {
		return code, err
	}
	if err := input.Required(); err != nil {
		return http.StatusBadRequest, err
	}
	return 0, nil
}

//	Parse request POST body as JSON array
func (a *Request) Request_JSON_slice(post_limit int, input any) (int, error, []error){
	b, code, err := a.request_JSON(post_limit)
	if err != nil {
		return code, err, nil
	}
	if err := json.Unmarshal(b, input, json.RejectUnknownMembers(true)); err != nil {
		switch t := err.(type) {
		case *json.SemanticError:
			serr, serrs := invalid_json.Slice_fields(t, b, input)
			return http.StatusBadRequest, serr, serrs
		}
		return http.StatusBadRequest, fmt.Errorf("Unable to unmarshal JSON"), nil
	}
	return 0, nil, nil
}

func (a *Request) Idempotency_hash() (string, error){
	s := a.r.Method+":"+a.r.RequestURI
	if header_type := a.r.Header.Get(head.CONTENT_TYPE); header_type != "" {
		s += ":"+header_type
	}
	if header_etag := a.r.Header.Get(head.IF_MATCH); header_etag != "" {
		s += ":"+header_etag
	}
	return hash.SHA256_hex(append([]byte(s+":"), a.body_received...)), nil
}

func (a *Request) request_JSON(post_limit int) ([]byte, int, error){
	b, code, err := a.Request(post_limit)
	if err != nil {
		return nil, code, err
	}
	if !head.Request_JSON(a.r) {
		return nil, http.StatusUnsupportedMediaType, fmt.Errorf("POST payload must be JSON")
	}
	return b, 0, nil
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
