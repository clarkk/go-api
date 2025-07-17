package api

import (
	"fmt"
	"log"
	"strings"
	"github.com/go-errors/errors"
	"github.com/clarkk/go-util/env"
)

//	Error JSON response and log unexpected error
func (a *Request) Error_log(code int, err error, e *env.Environment){
	a.Error(code, nil)
	var env_string string
	if e == nil {
		env_string = "<nil>"
	} else {
		env_string = fmt.Sprintf("%v", e.Data())
	}
	log.Printf("HTTP %d: %s %s %s\n\tEnv: %s\n\tPost payload (%d bytes): %s\n\n%s",
		code,
		a.r.Method,
		a.r.Host+a.r.URL.Path,
		a.r.URL.RawQuery,
		env_string,
		len(a.body_received),
		string(a.body_received),
		tab_indentation(errors.Wrap(err, 2).ErrorStack()),
	)
}

func tab_indentation(s string) string {
	return "\t"+strings.Join(strings.Split(s, "\n"), "\n\t")
}