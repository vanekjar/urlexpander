package urlexpander

import (
	"net/url"
	"time"

	"github.com/bluele/gcache"
)

type Config struct {
	// User agent string used when translating shortened url.
	UserAgent string

	// Maximum length of shortened url. It is assumed that no shortened url is longer than that.
	ShortUrlMaxLength int

	// Expanded urls are cached for repeated use. Using this option cache capacity can be set.
	CacheCapacity int

	// Expanded urls are cached for repeated use. Using this option expiration timeout can be set.
	CacheExpiration time.Duration
}

func newDefaultConfig() Config {
	return Config{
		UserAgent:         "Mozilla/5.0 (compatible; UrlExpander/1.0)",
		ShortUrlMaxLength: 32,
		CacheCapacity:     100000,
		CacheExpiration:   1 * time.Hour,
	}
}

type UrlExpander interface {
	// Expand given shortened url to its original form.
	// Return either an expanded url as a string or an error.
	ExpandUrl(shortenedUrl string) (string, error)
}

type expander struct {
	Config  Config
	fetcher *fetcher
	cache   gcache.Cache
}

// Create a new UrlExpander with default config
func New() UrlExpander {
	return NewFromConfig(newDefaultConfig())
}

// Create a new UrlExpander with provided config
func NewFromConfig(config Config) UrlExpander {
	return &expander{
		Config:  config,
		fetcher: newFetcher(config.UserAgent),
		cache:   gcache.New(config.CacheCapacity).LRU().Expiration(config.CacheExpiration).Build()}
}

func (exp expander) ExpandUrl(shortenedUrl string) (string, error) {
	if !exp.isShortened(shortenedUrl) {
		return "", ErrLongUrl
	}

	u, err := url.ParseRequestURI(shortenedUrl)
	if err != nil {
		return "", ErrInvalidUrl
	}

	// check if given url is present in cache
	r, err := exp.cache.GetIFPresent(u.String())
	if err == nil {
		// item is present in cache
		return r.(string), nil
	}

	expanded, err := exp.fetcher.fetchLocationHeader(u.String())
	if err != nil {
		return "", err
	}

	// set expanded url to cache
	exp.cache.Set(u.String(), expanded)

	return expanded, nil
}

// Check if the given string is not exceeding configured length limit.
func (exp expander) isShortened(u string) bool {
	return len(u) <= exp.Config.ShortUrlMaxLength
}
