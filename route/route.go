package routeapi

import (
	"errors"
	"strconv"
	"net/http"
	"github.com/clarkk/go-api"
	"github.com/clarkk/go-api/errin"
	"github.com/clarkk/go-api/etag"
	"github.com/clarkk/go-api/invalid_json"
	"github.com/clarkk/go-dbd"
	"github.com/clarkk/go-dtm"
	"github.com/clarkk/go-util/serv"
)

type Request struct {
	*api.Request
	r *http.Request
}

func New(w http.ResponseWriter, r *http.Request) *Request {
	return &Request{
		Request:	api.New(w, r),
		r:			r,
	}
}

func (a *Request) Allow_method(method serv.Method) bool {
	if a.Method() != string(method) {
		a.Header("Allow", string(method))
		a.Error(http.StatusMethodNotAllowed, nil)
		return false
	}
	return true
}

func (a *Request) Request_error(post_limit int, input api.Input) error {
	code, err := a.Payload_JSON_required(post_limit, input)
	if code == 0 {
		return nil
	}
	
	if t, ok := err.(*invalid_json.Type_error); ok {
		a.Errors(code, a.type_errors(t))
		return err
	}
	
	a.Error(code, err)
	return err
}

func (a *Request) Request_slice_error(post_limit int, input any) error {
	code, err, errs := a.Payload_JSON_slice(post_limit, input)
	if code == 0 {
		return nil
	}
	
	if err != nil {
		a.Error(code, err)
		return err
	}
	a.Bulk_semantic_errors(code, errs)
	return errors.New("semantic errors")
}

func (a *Request) Slug_id(seq int) (uint64, error){
	id_slug := serv.Get_slug(a.r, seq)
	if id_slug == "" {
		err := errors.New("ID is missing")
		a.Error(http.StatusBadRequest, err)
		return 0, err
	}
	return a.parse_slug_id(id_slug)
}

func (a *Request) Slug_id_optional(seq int) (uint64, error){
	id_slug := serv.Get_slug(a.r, seq)
	if id_slug == "" {
		return 0, nil
	}
	return a.parse_slug_id(id_slug)
}

func (a *Request) Match_etag(id uint64, m etag.Matcher) bool {
	code, err := etag.Match(a.Request, id, m)
	if code != 0 {
		a.Error(code, err)
		return false
	}
	if err != nil {
		a.Response_error(err)
		return false
	}
	return true
}

func (a *Request) Response_error(err error){
	a.Response_limit_error(err, nil)
}

func (a *Request) Response_limit_error(err error, limit *api.Limit){
	switch err {
	case dtm.ErrNotFound:
		a.Error(http.StatusNotFound, nil)
		return
	case dtm.ErrLimitOffsetRange:
		a.Error_limit(http.StatusBadRequest, err, limit)
		return
	}
	
	switch terr := err.(type) {
	case *dtm.Error:
		a.Error(terr.Code(), terr)
	case *dbd.Timeout_error:
		a.Error(http.StatusRequestTimeout, nil)
	default:
		a.Error_log(http.StatusInternalServerError, err, a.Env())
	}
}

func (a *Request) Entity_errors(errs errin.Map_lang, raw_error bool) errin.Map {
	var trans errin.Map
	errs.Each(func(field string, value *errin.Lang){
		var msg string
		if raw_error {
			msg = value.Key
		} else {
			msg = a.Env().Lang_error(value.Key, value.Replace).Error()
		}
		trans.Set(field, msg)
	})
	return trans
}

func (a *Request) Response_entity_errors(errs errin.Map_lang, raw_error bool){
	a.Errors(http.StatusUnprocessableEntity, a.Entity_errors(errs, raw_error))
}

func (a *Request) Response_entity_warnings(errs errin.Map_lang, raw_error bool){
	a.Warnings(http.StatusUnprocessableEntity, a.Entity_errors(errs, raw_error))
}

func (a *Request) type_errors(t *invalid_json.Type_error) errin.Map {
	var errs errin.Map
	
	for field, expect := range t.Expects() {
		var key string
		switch expect {
		case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "float32", "float64":
			key = "JSON_TYPE_NUMBER"
		case "string":
			key = "JSON_TYPE_STRING"
		default:
			key = "JSON_TYPE"
		}
		
		msg := a.Env().Lang_error(key, nil)
		errs.Set(field, msg.Error())
	}
	
	return errs
}

func (a *Request) parse_slug_id(id_slug string) (uint64, error){
	id, err := strconv.ParseUint(id_slug, 10, 64)
	if err != nil || id == 0 {
		err = errors.New("ID is not an integer")
		a.Error(http.StatusBadRequest, err)
		return 0, err
	}
	
	return id, nil
}