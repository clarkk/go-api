package api

import (
	"fmt"
	"strings"
	"reflect"
	"github.com/go-json-experiment/json"
)

func parse_unmarshal_json_error(err *json.SemanticError, b []byte, input any) error {
	required_fields, serr := required_input_fields(input)
	if serr != nil {
		return fmt.Errorf("Parse unmarshal JSON: %w", serr)
	}
	
	/*fmt.Printf("err: %v\n", err)
	fmt.Printf("ByteOffset: %d %T\n", err.ByteOffset, err.ByteOffset)
	fmt.Printf("JSONPointer: %s %T\n", err.JSONPointer, err.JSONPointer)
	fmt.Printf("JSONKind: %s %T\n", err.JSONKind, err.JSONKind)
	fmt.Printf("GoType: %s %T\n", err.GoType, err.GoType)
	fmt.Println("Err:", err.Err)*/
	
	//	Invalid/unknown fields
	if err.Err != nil {
		//if strings.HasPrefix(err.Err.Error(), "unknown name") {
			body_fields, serr := get_request_body_fields(b)
			if serr != nil {
				return serr
			}
			return fmt.Errorf("Invalid fields: %s", strings.Join(unknown_input_fields(body_fields, required_fields), ", "))
		//}
	} else {
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
		}
	}
	
	return fmt.Errorf("JSON unmarshal error at byte offset %d", err.ByteOffset)
}

func parse_unmarshal_json_slice_error(err *json.SemanticError, b []byte, input any) error {
	if _, serr := get_request_body_slice(b); serr != nil {
		return serr
	}
	return fmt.Errorf("JSON unmarshal error at byte offset %d", err.ByteOffset)
}

func required_input_fields(input any) (map[string]reflect.Type, error) {
	list := map[string]reflect.Type{}
	if reflect.TypeOf(input).Elem().Kind() != reflect.Struct {
		return list, fmt.Errorf("Input must be a struct")
	}
	struct_val := reflect.ValueOf(input).Elem().Type()
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

func get_request_body_fields(b []byte) (map[string]any, error) {
	var body any
	json.Unmarshal(b, &body)
	if reflect.TypeOf(body).Kind() != reflect.Map {
		return map[string]any{}, fmt.Errorf("Request body must be key-value pairs")
	}
	return body.(map[string]any), nil
}

func get_request_body_slice(b []byte) ([]any, error) {
	var body any
	json.Unmarshal(b, &body)
	if reflect.TypeOf(body).Kind() != reflect.Slice {
		return []any{}, fmt.Errorf("Request body must be an array")
	}
	return body.([]any), nil
}