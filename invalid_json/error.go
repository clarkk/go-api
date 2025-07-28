package invalid_json

import (
	"fmt"
	"strings"
)

type (
	Semantic_error struct {
		error 		string
		err 		error
	}
	
	Type_error struct {
		expects 	map[string]string
		err 		error
	}
)

func (e *Semantic_error) Error() string {
	return e.error
}

func (e *Semantic_error) Unwrap() error {
	return e.err
}

func (e *Type_error) Error() string {
	m := make([]string, len(e.expects))
	i := 0
	for field, expect := range e.expects {
		m[i] = fmt.Sprintf("Field %s expected data type '%s'", field, expect)
		i++
	}
	return strings.Join(m, ", ")
}

func (e *Type_error) Unwrap() error {
	return e.err
}

func (e *Type_error) Expects() map[string]string {
	return e.expects
}