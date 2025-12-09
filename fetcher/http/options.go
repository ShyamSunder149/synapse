package http

import "net/http"

type RequestOptions func(*http.Request)

func (f *HttpFetcher) WithBasicAuth(username, password string) RequestOptions {
	return func(req *http.Request) {
		req.SetBasicAuth(username, password)
	}
}

func (f *HttpFetcher) WithCookies(cookies []*http.Cookie) RequestOptions {
	return func(req *http.Request) {
		for _, cookie := range cookies {
			req.AddCookie(cookie)
		}
	}
}

func (f *HttpFetcher) WithHeaders(headers map[string]string) RequestOptions {
	return func(req *http.Request) {
		for k, v := range headers {
			req.Header.Add(k, v)
		}
	}
}

func (f *HttpFetcher) WithReferer(referer string) RequestOptions {
	return func(req *http.Request) {
		req.Header.Add("Referer", referer)
	}
}

func (f *HttpFetcher) WithUserAgent(userAgent string) RequestOptions {
	return func(req *http.Request) {
		req.Header.Add("User-Agent", userAgent)
	}
}
