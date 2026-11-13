package head

import (
	"time"
	"strconv"
	"strings"
	"net/http"
)

const (
	CONTENT_TYPE 		= "Content-Type"
	TYPE_JSON 			= "application/json"
	TYPE_JS 			= "application/javascript"
	//TYPE_FORM_DATA 		= "application/x-www-form-urlencoded"
	
	ACCEPT				= "Accept"
	
	ACCEPT_ENCODING 	= "Accept-Encoding"
	CONTENT_ENCODING 	= "Content-Encoding"
	ENCODING_GZIP 		= "gzip"
	
	//ACCEPT_LANG			= "Accept-Language"
	//CONTENT_LANG		= "Content-Language"
	
	USER_AGENT			= "User-Agent"
	VERSION 			= "Version"
	
	CACHE_CONTROL		= "Cache-Control"
	
	IF_MATCH			= "If-Match"
)

//	Check if a HTTP request is an API call or done via a browser
func Request_API(r *http.Request) bool {
	if r.Header.Get(USER_AGENT) == "" {
		return true
	}
	for _, v := range strings.Split(r.Header.Get(ACCEPT), ",") {
		v, _, _ = strings.Cut(v, ";")
		v = strings.TrimSpace(v)
		if v == "text/html" || v == "*/*" {
			return false
		}
	}
	return true
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
	return time.Unix(unix_time, 0).UTC().Format(http.TimeFormat)
}

func Response_allow_CORS(w http.ResponseWriter, allow string){
	w.Header().Set("Access-Control-Allow-Origin", allow)
}

func Response_cache_control_public_immutable(w http.ResponseWriter, days int){
	w.Header().Set(CACHE_CONTROL, "public, max-age="+strconv.Itoa(60 * 60 * 24 * days)+", immutable")
}