package invalid_json

import (
	"fmt"
	"strings"
	"reflect"
	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
)

func Fields(json_serr *json.SemanticError, b []byte, input any) error {
	body_fields, serr := request_fields(b)
	if serr != nil {
		return serr
	}
	
	input_fields := required_fields(input)
	
	if json_serr.Err != nil {
		if unknown_fields := unknown_request_fields(body_fields, input_fields); unknown_fields != nil {
			return &Semantic_error{"Invalid fields: "+strings.Join(unknown_fields, ", "), json_serr}
		}
	} else {
		serr := invalid_data_type(json_serr, b)
		if terr := invalid_fields_data_type(input_fields, body_fields); terr != nil {
			terr.err = serr
			return terr
		}
		if serr != nil {
			return serr
		}
	}
	
	return &Semantic_error{byte_offset_error(b, json_serr.ByteOffset), json_serr}
}

func Map_fields(json_serr *json.SemanticError, b []byte, inputs any) (error, map[string]error){
	_/*body_map*/, serr := request_map(b)
	if serr != nil {
		return serr, nil
	}
	
	rv := reflect.ValueOf(inputs).Elem()
	if rv.Kind() != reflect.Map {
		panic("Input must be a map")
	}
	
	//	TODO: add logic to return a readable parse error
	
	return &Semantic_error{byte_offset_error(b, json_serr.ByteOffset), json_serr}, nil
}

func Slice_fields(json_serr *json.SemanticError, b []byte, inputs any) (error, []error){
	body_slice, serr := request_slice(b)
	if serr != nil {
		return serr, nil
	}
	
	rv := reflect.ValueOf(inputs).Elem()
	if rv.Kind() != reflect.Slice {
		panic("Input must be a slice")
	}
	input := rv.Index(0)
	input_fields := required_fields_struct(input)
	
	has_errors	:= false
	errs		:= make([]error, len(body_slice))
	in			:= input.Interface()
	for i, b := range body_slice {
		body_fields, serr := request_fields(b)
		if serr != nil {
			has_errors = true
			errs[i] = serr
			continue
		}
		
		if unknown_fields := unknown_request_fields(body_fields, input_fields); unknown_fields != nil {
			has_errors = true
			errs[i] = &Semantic_error{"Invalid fields: "+strings.Join(unknown_fields, ", "), json_serr}
			continue
		}
		
		if entry_serr := json.Unmarshal(b, &in, json.RejectUnknownMembers(true)); entry_serr != nil {
			has_errors = true
			switch t := entry_serr.(type) {
			case *json.SemanticError:
				if serr := invalid_data_type(t, b); serr != nil {
					if terr := invalid_fields_data_type(input_fields, body_fields); terr != nil {
						terr.err = serr
						errs[i] = terr
						continue
					}
				}
				errs[i] = &Semantic_error{byte_offset_error(b, t.ByteOffset), t}
				continue
			}
			errs[i] = &Semantic_error{"Undefined error", entry_serr}
			continue
		}
	}
	
	if has_errors {
		return nil, errs
	}
	
	return &Semantic_error{byte_offset_error(b, json_serr.ByteOffset), json_serr}, nil
}

func invalid_fields_data_type(input_fields map[string]reflect.Type, body_fields map[string]any) *Type_error {
	err := &Type_error{}
	for field, rt := range input_fields {
		body_field, ok := body_fields[field]
		if ok {
			if rt.Kind() == reflect.Pointer {
				rt = rt.Elem()
			}
			if rt != reflect.TypeOf(body_field) {
				if err.expects == nil {
					err.expects = map[string]string{}
				}
				err.expects[field] = rt.String()
			}
		}
	}
	if err.expects == nil {
		return nil
	}
	return err
}

func invalid_data_type(err *json.SemanticError, b []byte) *Semantic_error {
	switch err.JSONKind {
	case 'n':
		return invalid_data_type_error(err, b, "null")
	case 'f', 't':
		return invalid_data_type_error(err, b, "bool")
	case '"':
		return invalid_data_type_error(err, b, "string")
	case '0':
		return invalid_data_type_error(err, b, "number")
	case '{', '}':
		return invalid_data_type_error(err, b, "object")
	case '[', ']':
		return invalid_data_type_error(err, b, "array")
	default:
		return nil
	}
}

func invalid_data_type_error(err *json.SemanticError, b []byte, kind string) *Semantic_error {
	return &Semantic_error{
		fmt.Sprintf("Invalid JSON type. Expected '%s' but got '%s' (at byte offset %d) after: %s", err.GoType, kind, err.ByteOffset, after_byte_offset(b, err.ByteOffset)),
		err,
	}
}

func request_fields(b []byte) (map[string]any, *Semantic_error){
	var body any
	if err := json.Unmarshal(b, &body); err != nil {
		return nil, &Semantic_error{"Request body must be key-value pairs", err}
	}
	fields, ok := body.(map[string]any)
	if !ok {
		return nil, &Semantic_error{"Request body must be key-value pairs", nil}
	}
	return fields, nil
}

func request_map(b []byte) (body map[string]jsontext.Value, serr *Semantic_error){
	if err := json.Unmarshal(b, &body); err != nil {
		serr = &Semantic_error{"Request body must be a map", err}
	}
	return
}

func request_slice(b []byte) (body []jsontext.Value, serr *Semantic_error){
	if err := json.Unmarshal(b, &body); err != nil {
		serr = &Semantic_error{"Request body must be an array", err}
	}
	return
}

func required_fields(input any) map[string]reflect.Type {
	fmt.Println(input)
	rv := reflect.ValueOf(input)
	for rv.Kind() == reflect.Pointer || rv.Kind() == reflect.Interface {
		rv = rv.Elem()
	}
	fmt.Println(rv)
	return required_fields_struct(rv)
}

func required_fields_struct(rv reflect.Value) map[string]reflect.Type {
	list := map[string]reflect.Type{}
	fmt.Println("kind:", rv.Kind())
	if rv.Kind() != reflect.Struct {
		panic("Input must be a struct")
	}
	rt := rv.Type()
	for i := range rt.NumField() {
		field := rt.Field(i)
		field_name := strings.Split(field.Tag.Get("json"), ",")[0]
		list[field_name] = field.Type
	}
	return list
}

func unknown_request_fields[K comparable, VT any, VR any](target map[K]VT, required map[K]VR) []K {
	var list []K
	for k := range target {
		if _, ok := required[k]; !ok {
			list = append(list, k)
		}
	}
	return list
}

func byte_offset_error(b []byte, byte_offset int64) string {
	return fmt.Sprintf("JSON unmarshal error (at byte offset %d) after: %s", byte_offset, after_byte_offset(b, byte_offset))
}

func after_byte_offset(b []byte, byte_offset int64) string {
	b = b[byte_offset:]
	n := 100
	if n < len(b) {
		b = append(b[:n], []byte{'.','.','.'}...)
	}
	return string(b)
}