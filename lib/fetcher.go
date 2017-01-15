package urlexpander

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/bluele/gcache"
	"github.com/temoto/robotstxt"
)

const robotsTxtCacheCapacity = 100
const robotsTxtCacheExpiration = 10 * time.Minute

var defaultHttpClient = &http.Client{
	Timeout: time.Second * 10,

	// disable following redirects
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

var robotsTxtPath, _ = url.Parse("/robots.txt")

type fetcher struct {
	httpClient     *http.Client
	userAgent      string
	robotsTxtCache gcache.Cache
}

func newFetcher(userAgent string) *fetcher {
	return newFetcherWithClient(defaultHttpClient, userAgent)
}

func newFetcherWithClient(httpClient *http.Client, userAgent string) *fetcher {
	return &fetcher{
		httpClient:     httpClient,
		userAgent:      userAgent,
		robotsTxtCache: gcache.New(robotsTxtCacheCapacity).LRU().Expiration(robotsTxtCacheExpiration).Build()}
}

// Fetch header of given url and return url that is found in "Location" header.
// If no "Location" header is present the original url is returned
func (f *fetcher) fetchLocationHeader(link string) (string, error) {
	allowed, err := f.isAllowedRobotsTxt(link)
	if err != nil {
		return "", err
	}
	if !allowed {
		return "", ErrDisallowedByRobotsTxt
	}

	req, err := http.NewRequest("HEAD", link, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", f.userAgent)

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	log.Printf("INFO: Requested %s, statusCode=%d", link, resp.StatusCode)

	if isRedirect(resp.StatusCode) {
		loc := resp.Header.Get("Location")
		if loc == "" {
			return "", fmt.Errorf("%d response missing Location header", resp.StatusCode)
		}

		u, err := req.URL.Parse(loc)
		if err != nil {
			return "", fmt.Errorf("failed to parse Location header %q: %v", loc, err)
		}

		return u.String(), nil
	}

	// if the original url is not redirect return it unmodified
	return link, nil
}

// Check if the given link is allowed to be crawled in robots.txt
func (f *fetcher) isAllowedRobotsTxt(link string) (bool, error) {
	u, err := url.ParseRequestURI(link)
	if err != nil {
		return false, err
	}

	robotsUrl := u.ResolveReference(robotsTxtPath)

	robots, err := f.fetchRobotsTxt(robotsUrl)
	if err != nil {
		return false, err
	}

	group := robots.FindGroup(f.userAgent)

	return group.Test(u.Path), nil
}

// Fetch and parse content of given robots.txt file
func (f *fetcher) fetchRobotsTxt(u *url.URL) (*robotstxt.RobotsData, error) {
	// check if given url is present in cache
	r, err := f.robotsTxtCache.GetIFPresent(u.String())
	if err == nil {
		// item is present in cache
		return r.(*robotstxt.RobotsData), nil
	}

	resp, err := f.httpClient.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	log.Printf("INFO: Requested %s, statusCode=%d", u.String(), resp.StatusCode)

	robots, err := robotstxt.FromResponse(resp)
	if err != nil {
		return nil, err
	}

	// set parsed robots.txt to cache
	f.robotsTxtCache.Set(u.String(), robots)

	return robots, nil
}

// True if the specified HTTP status code is redirect.
func isRedirect(statusCode int) bool {
	switch statusCode {
	case http.StatusMovedPermanently, http.StatusFound, http.StatusSeeOther,
		http.StatusTemporaryRedirect, http.StatusPermanentRedirect:
		return true
	}
	return false
}
