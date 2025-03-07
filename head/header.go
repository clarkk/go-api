package head

import (
	"time"
	"strings"
	"net/http"
)

const (
	CONTENT_TYPE 		= "Content-Type"
	TYPE_JSON 			= "application/json"
	//TYPE_FORM_DATA 		= "application/x-www-form-urlencoded"
	
	ACCEPT_ENCODING 	= "Accept-Encoding"
	CONTENT_ENCODING 	= "Content-Encoding"
	ENCODING_GZIP 		= "gzip"
	
	//ACCEPT_LANG			= "Accept-Language"
	//CONTENT_LANG		= "Content-Language"
	
	USER_AGENT			= "User-Agent"
	VERSION 			= "Version"
)

//	Check if a HTTP request is an API call or done via a browser
func Request_API(r *http.Request) bool {
	return r.Header.Get(USER_AGENT) == ""
}

//	Get API version
func Request_version(r *http.Request) string {
	return r.Header.Get(VERSION)
}

//	Check if HTTP request is JSON
func Request_JSON(r *http.Request) bool {
	content_type, _, _ := strings.Cut(r.Header.Get(CONTENT_TYPE), ";")
	return strings.TrimSpace(content_type) == TYPE_JSON
}

func GMT_unix_time(unix_time int64) string {
	return strings.Replace(time.Unix(unix_time, 0).Format(time.RFC1123), "UTC", "GMT", 1)
}