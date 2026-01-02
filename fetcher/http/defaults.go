package http

import (
	"net/http"
	"net/url"
)

var NoopEventHook = EventHooks{
	OnRequest:  func(*http.Request) {},
	OnResponse: func(*http.Response) {},
	OnError:    func(*http.Request, error) {},
	OnChunk:    func([]byte) {},
}

type NoopCookieJar struct{}

func (n *NoopCookieJar) SetCookies(u *url.URL, cookies []*http.Cookie) {}
func (n *NoopCookieJar) Cookies(u *url.URL) []*http.Cookie             { return nil }
