package urlexpander

import "testing"

import (
	"gopkg.in/jarcoal/httpmock.v1"
	"net/http"
)

func TestIsRedirect(t *testing.T) {
	redirectCodes := []int{301, 302, 303, 307, 308}

	for _, code := range redirectCodes {
		if !isRedirect(code) {
			t.Fatalf("%d should be redirect", code)
		}
	}
}

func TestIsAllowedRobotsTxt(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	robotsTxt := "User-agent: TestAgent\n"
	robotsTxt += "Disallow: /forbidden"

	httpmock.RegisterResponder("GET", "http://test.dev/robots.txt",
		httpmock.NewStringResponder(200, robotsTxt))

	fetcher := newFetcher("TestAgent")

	testCases := []struct {
		url     string
		allowed bool
	}{
		{url: "http://test.dev/forbidden",
			allowed: false,
		},
		{url: "http://test.dev/index",
			allowed: true,
		},
	}

	for _, testCase := range testCases {
		allowed, err := fetcher.isAllowedRobotsTxt(testCase.url)
		if err != nil {
			t.Fatal(err)
		}
		if allowed != testCase.allowed {
			t.Fatalf("Url %s should be allowed=%t", testCase.url, testCase.allowed)
		}
	}

	// change userAgent of fetcher
	fetcher.userAgent = "UrlExpander"

	// now all request should be allowed
	for _, testCase := range testCases {
		allowed, err := fetcher.isAllowedRobotsTxt(testCase.url)
		if err != nil {
			t.Fatal(err)
		}
		if allowed != true {
			t.Fatalf("Url %s should be allowed", testCase.url)
		}
	}
}

func TestFetchLocationHeader(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("HEAD", "http://test.dev/redirected",
		func(req *http.Request) (*http.Response, error) {
			resp := httpmock.NewStringResponse(302, "")
			resp.Header.Add("Location", "/target")
			return resp, nil
		})

	httpmock.RegisterResponder("HEAD", "http://test.dev/not-redirected",
		httpmock.NewStringResponder(200, ""))

	httpmock.RegisterResponder("GET", "http://test.dev/robots.txt",
		httpmock.NewStringResponder(200, ""))

	fetcher := newFetcher("TestAgent")

	target, err := fetcher.fetchLocationHeader("http://test.dev/redirected")
	if err != nil {
		t.Fatal(err)
	}
	if target != "http://test.dev/target" {
		t.Fatal("Url http://test.dev/redirected should be redirected")
	}

	target, err = fetcher.fetchLocationHeader("http://test.dev/not-redirected")
	if err != nil {
		t.Fatal(err)
	}
	if target != "http://test.dev/not-redirected" {
		t.Fatal("Url http://test.dev/not-redirected shouldn't be redirected")
	}
}
