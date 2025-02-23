package api

import (
	"fmt"
	"strings"
	"reflect"
	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
)

type Semantic_error struct {
	error 	string
	err 	error
}

func (e *Semantic_error) Error() string {
	return e.error
}

func (e *Semantic_error) Unwrap() error {
	return e.err
}

func parse_unmarshal_json_slice_error(err *json.SemanticError, b []byte, inputs any) (error, []error){
	if err.Err != nil {
		body_slice, serr := get_request_body_slice(b)
		if serr != nil {
			return serr, nil
		}
		
		elem := reflect.ValueOf(inputs).Elem()
		if elem.Kind() != reflect.Slice {
			return &Semantic_error{"Input must be a slice", err}, nil
		}
		
		input := elem.Index(0)
		required_fields, serr := required_input_fields_elem(input)
		if serr != nil {
			return serr, nil
		}
		
		has_errors	:= false
		errs		:= make([]error, len(body_slice))
		in			:= input.Interface()
		for i, b := range body_slice {
			body_fields, serr := get_request_body_fields(b)
			if serr != nil {
				has_errors = true
				errs[i] = serr
				continue
			}
			
			unknown_fields := unknown_input_fields(body_fields, required_fields)
			if len(unknown_fields) != 0 {
				has_errors = true
				errs[i] = &Semantic_error{fmt.Sprintf("Invalid fields: %s", strings.Join(unknown_fields, ", ")), err}
				continue
			}
			
			if serr := json.Unmarshal(b, &in, json.RejectUnknownMembers(true)); serr != nil {
				has_errors = true
				errs[i] = &Semantic_error{"Undefined error", serr}
				continue
			}
		}
		
		if has_errors {
			return nil, errs
		}
	} else {
		if terr := invalid_data_type(err, b); terr != nil {
			return terr, nil
		}
	}
	
	return &Semantic_error{parse_unmarshal_json_byte_offset(b, err.ByteOffset), err}, nil
}

func parse_unmarshal_json_byte_offset(b []byte, byte_offset int64) string {
	return fmt.Sprintf("JSON unmarshal error (at byte offset %d) after: %s", byte_offset, after_byte_offset(b, byte_offset))
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

func after_byte_offset(b []byte, byte_offset int64) string {
	b = b[byte_offset:]
	n := 100
	if n < len(b) {
		b = append(b[:n], []byte{'.','.','.'}...)
	}
	return string(b)
}

func required_input_fields_elem(elem reflect.Value) (map[string]reflect.Type, *Semantic_error){
	list := map[string]reflect.Type{}
	if elem.Kind() != reflect.Struct {
		return nil, &Semantic_error{"Input must be a struct", nil}
	}
	struct_val := elem.Type()
	for i := 0; i < struct_val.NumField(); i++ {
		field := struct_val.Field(i)
		list[strings.ToLower(field.Name)] = field.Type
	}
	return list, nil
}

func unknown_input_fields[K comparable, VT any, VR any](target map[K]VT, required map[K]VR) []K {
	list := []K{}
	for k := range target {
		if _, ok := required[k]; !ok {
			list = append(list, k)
		}
	}
	return list
}

func get_request_body_fields(b []byte) (map[string]any, *Semantic_error){
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