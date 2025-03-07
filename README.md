# Install
`go get -u github.com/clarkk/go-api`

- [go-api](#go-api) Lightweight JSON API for HTTP server
- [go-api/errin](#go-apierrin) Simple request validation error handling
- [go-api/idem](#go-apiidem) Lightweight API idempotency cache (width Redis)

# go-api
Lightweight JSON API for HTTP server with idempotency handling and entity-tag (ETag) to identify a specific version of a resource.

```
package main

import (
  "fmt"
  "errors"
  "net/http"
  "github.com/clarkk/go-api"
  "github.com/clarkk/go-util/serv"
)

type json_input struct {
  Name      *string   `json:"name"`
  Email     *string   `json:"email"`
}

func main(){
  h := serv.NewHTTP("domain.com", "127.0.0.1", 8000)
  
  h.Subhost("subdomain.").
    Route_exact(serv.POST, "/create", 60, func(w http.ResponseWriter, r *http.Request){
      defer serv.Recover(w)
      
      /*
        Handle GZIP encoding or let reverse proxy (nginx) handle it
        true = handle GZIP
        false = let reverse proxy (nginx) handle it
      */
      handle_gzip = false
      
      a := api.New(w, r, handle_gzip)
      defer a.Recover()
      
      //  Max request post size in kb
      post_limit := 1024
      
      //  Parse JSON into struct
      var input json_input
      if code, err := a.Request_JSON(post_limit, &input); code != 0 {
        a.Error(code, err)
        return
      }
      
      fmt.Println(*input.Name, *input.Email)
      
      something_went_wrong := true
      if something_went_wrong {
        a.Errorf(http.StatusBadRequest, "Something went wrong: %s", "Bad thing")
        return
      }
      
      something_went_wrong_again := true
      if something_went_wrong_again {
        a.Error(http.StatusBadRequest, errors.New("Failed!"))
        return
      }
      
      //  Process some business logic here
      res := api.Response_result{
        Result: map[string]any{
          "success": true,
        },
      }
      
      a.Response_JSON(http.StatusOK, res)
    })
   
  h.Run()
}
```

# go-api/errin
Simple way to handle multiple input validating errors simultaneously and return all errors to the client

```
package main

import (
  "fmt"
  "github.com/clarkk/go-api/errin"
)

func main(){
  if errs := validate_input(); errs != nil {
    //  Validating failed
    fmt.Println(errs)
  }
}

func validate_input() errin.Map {
  var errs errin.Map
  
  errs.Set("name", "Name is invalid")
  errs.Set("email", "E-mail is invalid")
  
  return errs
}
```

# go-api/idem
Lightweight API idempotency cache (width Redis)
- Caches responses via Redis
- Ensures duplicate HTTP POST requests etc. will not create duplicate entries in the database

Idempotency is a property of certain operations or API requests, which guarantees that performing the operation multiple times will yield the same result as if it was executed only once.
If a network error occurs and the response is never received by the client, it is possible to call the HTTP request again with an identical `Idempotency-Key` to receive the lost response.
It therefore only makes sense to implement idempotency on writing operations and never on reading operations.

### Example
When idempotency is initiated in a HTTP handler a `Idempotency-Key` header with a unique identifier (e.g. `ULID` or `UUID`) is required to request the resource.
If the HTTP response returns a cached result the `Idempotency-Key-Cached` header contains the cached timestamp.

## HTTP request
```
POST /create HTTP/2
...
Idempotency-Key: a-unique-identifier-for-each-request
```

## HTTP server
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
  //  Required to store/cache idempotency responses
  rdb.Connect("127.0.0.1", 6379, "redis-auth")
  
  h := serv.NewHTTP("domain.com", "127.0.0.1", 8000)
  
  h.Subhost("subdomain.").
    Route_exact(serv.POST, "/create", 60, func(w http.ResponseWriter, r *http.Request){
      defer serv.Recover(w)
      
      //  Handle GZIP encoding or let reverse proxy (nginx) handle it
      handle_gzip = false
      
      a := api.New(w, r, handle_gzip)
      defer a.Recover()
      
      //  Set a unique identifier for the user or session to avoid duplicate idempotency keys
      //  accros multiple users. Could be a user-id
      uid := "unique-user-or-session-identifier"
      
      //  Set idempotency header to required or optional
      //  If optional only responses with a idempotency header is cached
      idempotency_required := false
      
      //  Initiates idempotency
      idempotency, err := idem.New(a, uid, idempotency_required)
      //  Get cached idempotency response or return error response
      if err != nil || idempotency.Cached() {
        return
      }
      
      //  Process some business logic here
      res := api.Response_result{
        Result: map[string]any{
          "success": true,
        },
      }
      
      //  Cache idempotency response and send response
      idempotency.Response_JSON(http.StatusOK, res)
    })
  
  h.Run()
}
```