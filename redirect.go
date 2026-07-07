package api

import (
	"net/http"
	"github.com/clarkk/go-api/head"
)

//	Redirect without caching
func (a *Request) Redirect(status int, url string){
	if a.w.Sent_header() {
		panic("HTTP header already sent. Can not redirect to: "+url)
	}
	a.w.Header().Set(head.CACHE_CONTROL, "no-store")
	http.Redirect(a.w, a.r, url, status)
}

//	Redirect with caching
func (a *Request) Redirect_cache(status int, url string){
	if a.w.Sent_header() {
		panic("HTTP header already sent. Can not redirect to: "+url)
	}
	http.Redirect(a.w, a.r, url, status)
}