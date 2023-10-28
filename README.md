# Install
`go get -u github.com/clarkk/go-api`

# go-api
Lightweight HTTP API middleware to HTTP server

# go-api/idem
Lightweight API idempotency cache
- Caches responses via Redis
- Ensures that duplicate HTTP POST requests will not create duplicate entries in the database

Idempotency is a property of certain operations or API requests, which guarantees that performing the operation multiple times will yield the same result as if it was executed only once.

### Example
```
package main

import (
  "io"
  "net/http"
  "github.com/clarkk/go-api/idem"
  "github.com/clarkk/go-util/rdb"
)

func main(){
  rdb.Connect(REDIS_AUTH)
  
  h.Route(serv.POST, "/create", 60, func(w http.ResponseWriter, r *http.Request){
    //  Initiate idempotency
    idempotency := idem.Init(r, "a-unique-user-or-session-identifier")
    if code, err := idempotency.Error(); err != nil {
      a.Error(code, err)
      return
    }
    
    //  Response idempotency cache
    if code, res, header := idempotency.Cache(); code != 0 {
      a.Header(header.Key, header.Value)
      a.Response(code, res)
      return
    }
    
    //  Cache idempotency response
    go idempotency.Store_JSON(http.StatusOK, res)
    
    a.Response_JSON(code, res)
  })
}
```