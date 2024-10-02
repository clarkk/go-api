package api

import (
	"fmt"
	"strings"
	"reflect"
	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
)

func parse_unmarshal_json_error(err *json.SemanticError, b []byte, input any) error {
	/*fmt.Printf("err: %v\n", err)
	fmt.Printf("ByteOffset: %d %T\n", err.ByteOffset, err.ByteOffset)
	fmt.Printf("JSONPointer: %s %T\n", err.JSONPointer, err.JSONPointer)
	fmt.Printf("JSONKind: %s %T\n", err.JSONKind, err.JSONKind)
	fmt.Printf("GoType: %s %T\n", err.GoType, err.GoType)
	fmt.Println("Err:", err.Err)*/
	
	if err.Err != nil {
		body_fields, serr := get_request_body_fields(b)
		if serr != nil {
			return serr
		}
		
		required_fields, serr := required_input_fields(input)
		if serr != nil {
			return serr
		}
		
		unknown_fields := unknown_input_fields(body_fields, required_fields)
		if len(unknown_fields) != 0 {
			return fmt.Errorf("Invalid fields: %s", strings.Join(unknown_fields, ", "))
		}
	} else {
		if terr := invalid_field_type(err); terr != nil {
			return terr
		}
	}
	
	return fmt.Errorf("JSON unmarshal error at byte offset %d", err.ByteOffset)
}

func parse_unmarshal_json_slice_error(err *json.SemanticError, b []byte, inputs any) (error, []error){
	if err.Err != nil {
		body_slice, serr := get_request_body_slice(b)
		if serr != nil {
			return serr, nil
		}
		
		elem := reflect.ValueOf(inputs).Elem()
		if elem.Kind() != reflect.Slice {
			return fmt.Errorf("Input must be a slice"), nil
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
				errs[i] = fmt.Errorf("Invalid fields: %s", strings.Join(unknown_fields, ", "))
				continue
			}
			
			if serr := json.Unmarshal(b, &in, json.RejectUnknownMembers(true)); err != nil {
				has_errors = true
				errs[i] = serr
				continue
			}
		}
		
		if has_errors {
			return nil, errs
		}
	} else {
		if terr := invalid_field_type(err); terr != nil {
			return terr, nil
		}
	}
	
	return fmt.Errorf("JSON unmarshal error at byte offset %d", err.ByteOffset), nil
}

func invalid_field_type(err *json.SemanticError) error {
	//	Invalid type
	switch err.JSONKind {
	case 'n':
		return fmt.Errorf("Invalid JSON type NULL to %s", err.GoType)
	case 'f', 't':
		return fmt.Errorf("Invalid JSON type bool to %s", err.GoType)
	case '"':
		return fmt.Errorf("Invalid JSON type string to %s", err.GoType)
	case '0':
		return fmt.Errorf("Invalid JSON type number to %s", err.GoType)
	case '{', '}':
		return fmt.Errorf("Invalid JSON type object to %s", err.GoType)
	case '[', ']':
		return fmt.Errorf("Invalid JSON type array to %s", err.GoType)
	default:
		return nil
	}
}

func required_input_fields(input any) (map[string]reflect.Type, error){
	return required_input_fields_elem(reflect.ValueOf(input).Elem())
}

func required_input_fields_elem(elem reflect.Value) (map[string]reflect.Type, error){
	list := map[string]reflect.Type{}
	if elem.Kind() != reflect.Struct {
		return list, fmt.Errorf("Input must be a struct")
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

func get_request_body_fields(b []byte) (map[string]any, error){
	var body any
	if err := json.Unmarshal(b, &body); err != nil {
		return nil, fmt.Errorf("Request body must be key-value pairs")
	}
	out, ok := body.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("Request body must be key-value pairs")
	}
	return out, nil
}

func get_request_body_slice(b []byte) ([]jsontext.Value, error){
	var body []jsontext.Value
	if err := json.Unmarshal(b, &body); err != nil {
		return nil, fmt.Errorf("Request body must be an array")
	}
	return body, nil
}