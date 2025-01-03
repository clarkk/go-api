package api

import (
	"fmt"
	"log"
	"context"
	"strings"
	"net/http"
	"path/filepath"
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
	
	List map[string]string
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