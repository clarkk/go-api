# Install
`go get -u github.com/clarkk/go-api`

# go-api
Lightweight JSON API middleware to HTTP server

# go-api/idem
Lightweight API idempotency cache (width Redis)
- Caches responses via Redis
- Ensures duplicate HTTP POST requests will not create duplicate entries in the database

Idempotency is a property of certain operations or API requests, which guarantees that performing the operation multiple times will yield the same result as if it was executed only once.

### Example
When idempotency is initiated the HTTP request must provide a unique `Idempotency-Key` header value.
```
POST /create HTTP/2
...
Idempotency-Key: a-unique-identifier-for-each-request
```

```
package main

import (
  "net/http"
  "github.com/clarkk/go-api"
  "github.com/clarkk/go-api/idem"
  "github.com/clarkk/go-util/rdb"
  "github.com/clarkk/go-util/serv"
)

func main(){
  //  Required to store idempotency responses
  rdb.Connect(REDIS_AUTH)
  
  h := serv.NewHTTP("127.0.0.1", 3000)
  
  h.Route(serv.POST, "/create", TIMEOUT, func(w http.ResponseWriter, r *http.Request){
    defer serv.Recover(w)
    
    a := api.NewRequest(w, r)
    defer a.Recover()
    
    //  Initiates idempotency
    idempotency := idem.Init(r, "a-unique-user-or-session-identifier")
    if code, err := idempotency.Error(); err != nil {
      a.Error(code, err)
      return
    }
    
    //  Get cached idempotency response
    if code, res, header := idempotency.Cache(); code != 0 {
      a.Header(header.Key, header.Value)
      a.Response_code(code, res)
      return
    }
    
    //  Process some business logic here
    res := map[string]string{
      "success": "ok"
    }
    
    //  Cache idempotency response
    go idempotency.Store_JSON(http.StatusOK, res)
    
    a.Response_JSON(res)
  })
  
  h.Run()
}
```

## Wrap API around `http.HandlerFunc`
```
package main

import (
  "net/http"
  "github.com/clarkk/go-api"
  "github.com/clarkk/go-util/rdb"
  "github.com/clarkk/go-util/serv"
)

func wrap_api(h http.HandlerFunc) http.HandlerFunc {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
    defer serv.Recover(w)
    
    //  Initiate new API request
    a := api.NewRequest(w, r)
    defer a.Recover()
    
    //  Verify authentication here
    if err := do_some_authentication(); err != nil {
      a.Error(http.StatusUnauthorized, err)
      return
    }
    
    //  Serve the wrapped handler
    h.ServeHTTP(w, a.Wrap())
  })
}

func main(){
  h := serv.NewHTTP("127.0.0.1", 3000)
  
  h.Route_regex(serv.GET, "/get/([^/]+)", TIMEOUT, wrap_api(func(w http.ResponseWriter, r *http.Request){
    //  Fetch the API from wrapper
    a := api.Wrapped(r)
    
    table := serv.Get_pattern_slug(r, 0)
    
    res := map[string]string{
      "table": table,
    }
    
    //  Response
    a.Response_JSON(res)
  }))
  
  h.Run()
}
```