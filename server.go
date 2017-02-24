package main

import (
	urlexpander "./lib"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"syscall"
	"time"
)

// parse flags from cmd-line
func parseArguments() (port int, userAgent string, cacheCapacity int, cacheExpiration time.Duration, apiOnly bool) {
	flag.BoolVar(&apiOnly, "api-only", false, "Expose only the API end points")
	flag.IntVar(&port, "port", 8080, "Bind webserver to given port.")
	flag.StringVar(&userAgent, "user-agent",
		"Mozilla/5.0 (compatible; UrlExpander/1.0)",
		" User agent string used when translating shortened url.")
	flag.IntVar(&cacheCapacity, "cache-capacity",
		100000,
		"Expanded urls are cached for repeated queries. Using this option cache capacity can be set.")

	var cacheExpirationMinutes int
	flag.IntVar(&cacheExpirationMinutes, "cache-expiration",
		60,
		"Set cache expiration time in minutes.")

	help := flag.Bool("help", false, "Print usage.")
	flag.Parse()

	cacheExpiration = time.Duration(cacheExpirationMinutes) * time.Minute

	// print usage
	if *help {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		syscall.Exit(0)
	}

	return
}

type App struct {
	Expander urlexpander.UrlExpander
}

func main() {
	port, userAgent, cacheCapacity, cacheExpiration, apiOnly := parseArguments()

	conf := urlexpander.Config{
		CacheCapacity:     cacheCapacity,
		CacheExpiration:   cacheExpiration,
		UserAgent:         userAgent,
		ShortUrlMaxLength: 32,
	}

	log.Printf("INFO: Used configuration: %#v", conf)

	app := App{Expander: urlexpander.NewFromConfig(conf)}

	log.Printf("INFO: Listening on port %d", port)
	if !apiOnly {
		http.Handle("/", http.FileServer(http.Dir("./static")))
	}
	http.HandleFunc("/api/expand", app.expand)
	err := http.ListenAndServe(":"+strconv.Itoa(port), nil)
	if err != nil {
		panic(err)
	}
}

type responseMessage struct {
	Original string `json:"original,omitempty"`
	Expanded string `json:"expanded,omitempty"`
	Error    string `json:"error,omitempty"`
}

// Handle /expand API endpoint calls
func (app *App) expand(w http.ResponseWriter, r *http.Request) {
	shortened := r.URL.Query().Get("url")
	if shortened == "" {
		resp, _ := json.Marshal(responseMessage{Error: "No url provided. Use 'url' param."})
		http.Error(w, string(resp), http.StatusBadRequest)
		return
	}

	expanded, err := app.Expander.ExpandUrl(shortened)
	if err != nil {
		log.Printf("ERROR: (%s) %s", r.RemoteAddr, err.Error())
		switch err {
		case urlexpander.ErrDisallowedByRobotsTxt, urlexpander.ErrInvalidUrl, urlexpander.ErrLongUrl:
			resp, _ := json.Marshal(responseMessage{Error: err.Error()})
			http.Error(w, string(resp), http.StatusBadRequest)
		default:
			resp, _ := json.Marshal(responseMessage{Error: "Uknown error"})
			http.Error(w, string(resp), http.StatusInternalServerError)
		}
		return
	}

	resp := responseMessage{Original: shortened, Expanded: expanded}
	json.NewEncoder(w).Encode(resp)
}
