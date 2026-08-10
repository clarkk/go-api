# Install
`go get -u github.com/clarkk/go-api`

- [go-api](#go-api) Lightweight JSON API for HTTP server
- [go-api/errin](#go-apierrin) Simple request validation error handling
- [go-api/idem](#go-apiidem) Lightweight API idempotency cache (width Redis)
- [go-api/map_json](#go-apimap_json) Ordered maps in JSON responses

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
      
      a := api.New(w, r)
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
Simple error maps for collecting and returning multiple input validation errors at once.

Instead of returning on the first validation error, `errin` lets you collect errors for multiple fields and return them together.

## Basic usage
```
package main

import (
  "fmt"
  "github.com/clarkk/go-api/errin"
)

func main() {
  if errs := validate_input(); errs != nil {
    fmt.Println(errs)
  }
}

func validate_input() errin.Map {
  var errs errin.Map
  
  errs.Set("name", "Name is invalid")
  errs.Set("email", "E-mail is invalid")
  
  if len(errs) == 0 {
    return nil
  }
  
  return errs
}
```

A `Map` stores one error message per key. Calling `Set` with an existing key replaces its value.

## Checking for errors

Use `Has` to check whether a specific key has an error:
```
if errs.Has("email") {
  // Email validation failed
}
```

Use `Each` to iterate over all errors:
```
errs.Each(func(key, value string) {
  fmt.Println(key, value)
})
```

## JSON output

`Output` converts the error map to a `map_json.Map`, an ordered map that preserves insertion order, making it suitable for use in an API response.
```
output := errs.Output()
```

An empty map returns `nil`.

Example:
```
errs.Set("name", "Name is invalid")
errs.Set("email", "E-mail is invalid")

fmt.Println(errs.Output())
```

Output:
```
{
  "name": "Name is invalid",
  "email": "E-mail is invalid"
}
```

## String representation

`Map` implements `fmt.Stringer`:
```
fmt.Println(errs)
```

Output:
```
name: Name is invalid, email: E-mail is invalid
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
      
      a := api.New(w, r)
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

# go-api/map_json

Ordered JSON objects.

`map_json.Map` provides a simple key/value structure that behaves like a JSON object while preserving the order in which keys are added.

Unlike a regular Go map, `map_json.Map` guarantees deterministic key ordering when marshaling to JSON.

## Basic usage

```
package main

import (
  "fmt"
  
  "github.com/clarkk/go-api/map_json"
)

func main() {
  m := map_json.New()
  
  m.Set("name", "John")
  m.Set("email", "john@example.com")
  m.Set("age", 42)
  
  data, err := m.MarshalJSON()
  if err != nil {
    panic(err)
  }
  
  fmt.Println(string(data))
}
```

Output:
```
{
  "name": "John",
  "email": "john@example.com",
  "age": 42
}
```

The keys are serialized in the same order they were added.

## Setting values

Values can be any Go value that can be encoded as JSON:
```
m.Set("string", "value")
m.Set("number", 123)
m.Set("boolean", true)
m.Set("null", nil)
m.Set("array", []string{"one", "two"})
m.Set("object", map[string]any{
  "foo": "bar",
})
```

Calling `Set` with an existing key updates its value **without** changing its position:
```
m.Set("name", "John")
m.Set("email", "john@example.com")
m.Set("name", "Jane")
```

The resulting order remains:
```
name, email
```

## Getting values

Use `Get` to retrieve a value:
```
value, ok := m.Get("name")
if ok {
  fmt.Println(value)
}
```

`Get` returns `(nil, false)` when the key does not exist.

It is also safe to call `Get` on a nil `*Map`:
```
var m *map_json.Map
value, ok := m.Get("name")

// value == nil
// ok == false
```

## Keys

Use `Keys` to get all keys in insertion order:
```
keys := m.Keys()

fmt.Println(keys)
```

Output:
```
[name email age]
```

For a nil map, `Keys` returns `nil`.

## Length

Use `Len` to get the number of entries:
```
fmt.Println(m.Len())
```

A nil map returns `0`.

## JSON marshaling

`Map` implements `json.Marshaler`, so it can be passed directly to `encoding/json`:
```
data, err := json.Marshal(m)
```

It can also be embedded in another structure:
```
type Response struct {
  Data *map_json.Map `json:"data"`
}

response := Response{
  Data: m,
}

data, err := json.Marshal(response)
```

The resulting JSON preserves the insertion order of the keys in `Map`.