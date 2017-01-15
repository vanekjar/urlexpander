# UrlExpander

[![GoDoc](https://godoc.org/github.com/vanekjar/urlexpander/lib?status.svg)](https://godoc.org/github.com/vanekjar/urlexpander/lib)

http://urlexpander.tk

Go package providing API for expanding shortened urls from services like _goo.gl, bitly.com, tinyurl.com_

## Features

 * Translates shortened urls as fast as possible by sending lightweight HEAD request to shortening service
 * Uses local cache to handle repeated queries
 * Respects _robots.txt_ in case the shortening service must not be visited by crawlers

## Usage

This project can be used either as a library from Go code or it can be used as a standalone service providing HTTP API.
  
### Library
  
```go
import "github.com/vanekjar/urlexpander/lib"

expander := urlexpander.New()
expanded, err := expander.ExpandUrl("https://goo.gl/HFoP0a")
```

### HTTP API server

Install __UrlExpander__ locally
 
```
go get github.com/vanekjar/urlexpander
```

Run command that will start a local HTTP server (listening on port 8080 by default)

```bash
urlexpander
```

Check running server by visiting http://localhost:8080

#### API description

##### Request

```
GET http://urlexpander.tk/api/expand?url=https://goo.gl/HFoP0a
```

##### Response

```
200 OK
{"original":"https://goo.gl/HFoP0a", "expanded":"http://urlexpander.tk"} 
```

## Configuration

```go
conf := urlexpander.Config{
  // Expanded urls are cached for repeated queries. Set cache capacity.
  CacheCapacity:     cacheCapacity,
  // Set cache expiration time in minutes.
  CacheExpiration:   cacheExpiration,
  // User agent string used when translating shortened url.
  UserAgent:         userAgent,
  // Maximum length of shortened url. It is assumed that no shortened url is longer than that.
  ShortUrlMaxLength: 32,
}

expander := urlexpander.NewFromConfig(conf)
```