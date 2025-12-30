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