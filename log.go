package api

import (
	"fmt"
	"log"
	"strings"
	"github.com/go-errors/errors"
	"github.com/go-json-experiment/json"
	"github.com/clarkk/go-util/env"
)

//	Error JSON response and log unexpected error
func (a *Request) Error_log(code int, err error, e *env.Environment){
	a.Error(code, nil)
	a.log(code, err, e)
}

func (a *Request) log(code int, err error, e *env.Environment){
	env_string := "<nil>"
	if e != nil {
		if b, err := json.Marshal(e.Data()); err == nil {
			env_string = string(b)
		} else {
			env_string = fmt.Sprintf("%v", e.Data())
		}
	}
	s := fmt.Sprintf("HTTP %d: %s %s %s\nEnv: %s\nPost payload (%d bytes): %s\n\n%s",
		code,
		a.r.Method,
		a.r.Host+a.r.URL.Path,
		a.r.URL.RawQuery,
		env_string,
		len(a.body_received),
		a.body_received,
		errors.Wrap(err, 2).ErrorStack(),
	)
	log.Printf(tab_indentation(s))
}

func tab_indentation(s string) string {
	if s == "" {
		return ""
	}
	return strings.ReplaceAll(s, "\n", "\n\t")
}