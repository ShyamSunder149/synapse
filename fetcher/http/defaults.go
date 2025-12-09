package http

import "net/http"

var NoopEventHook = EventHooks{
	OnRequest:  func(*http.Request) {},
	OnResponse: func(*http.Response) {},
	OnError:    func(*http.Request, error) {},
	OnChunk:    func([]byte) {},
}
