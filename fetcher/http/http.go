package http

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"

	"github.com/ritvikos/synapse/policy"
	"golang.org/x/net/publicsuffix"
)

type HttpFetcher struct {
	httpClient      HttpClient
	retryController policy.RetryPolicy
	eventHook       EventHooks
	cookieJar       http.CookieJar
}

type Options func(*HttpFetcher)

func NewHttpFetcher(httpClient HttpClient, opts ...Options) HttpFetcher {
	jar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})

	fetcher := HttpFetcher{
		httpClient: httpClient,
		eventHook:  NoopEventHook,
		cookieJar:  jar,
	}

	for _, opt := range opts {
		opt(&fetcher)
	}

	return fetcher
}

func WithEventHooks(hooks EventHooks) Options {
	return func(f *HttpFetcher) {
		f.eventHook = hooks
	}
}

func WithCookieJar(jar http.CookieJar) Options {
	return func(f *HttpFetcher) {
		if jar == nil {
			f.cookieJar = &NoopCookieJar{}
			return
		}
		f.cookieJar = jar
	}
}

func (f *HttpFetcher) Head(ctx context.Context, url string, opts ...RequestOptions) (*http.Response, error) {
	return f.doRequest(ctx, http.MethodHead, url, nil, opts...)
}

func (f *HttpFetcher) Get(ctx context.Context, url string, opts ...RequestOptions) (*http.Response, error) {
	return f.doRequest(ctx, http.MethodGet, url, nil)
}

// -- POST --
// func (f *HttpFetcher) PostForm(ctx context.Context, endpoint string, data map[string]string, opts ...RequestOptions) (*http.Response, error) {
// 	formData := make(url.Values)
// 	for key, value := range data {
// 		formData.Set(key, value)
// 	}
// 	body := strings.NewReader(formData.Encode())

// 	// Add Content-Type header for form data
// 	opts = append(opts, func(req *http.Request) {
// 		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
// 	})

// 	return f.doRequest(ctx, http.MethodPost, endpoint, body, opts...)
// }

// func (f *HttpFetcher) PostRaw(ctx context.Context, url string, data []byte, opts ...RequestOptions) (*http.Response, error) {
// 	body := bytes.NewBuffer(data)
// 	return f.doRequest(ctx, http.MethodPost, url, body, opts...)
// }

// func (f *HttpFetcher) PostMultipart(ctx context.Context, url string, data map[string][]byte) error {}

func (f *HttpFetcher) doRequest(ctx context.Context, method string, url string, body io.Reader, opts ...RequestOptions) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(req)
	}

	return f.do(ctx, req)
}

func (f *HttpFetcher) do(_ context.Context, req *http.Request) (*http.Response, error) {
	for _, cookie := range f.cookieJar.Cookies(req.URL) {
		req.AddCookie(cookie)
	}

	f.eventHook.OnRequest(req)

	resp, err := f.httpClient.Do(req)
	if err != nil {
		f.eventHook.OnError(req, err)
		return nil, err
	}

	if cookies := resp.Cookies(); len(cookies) > 0 {
		f.cookieJar.SetCookies(req.URL, cookies)
	}

	f.eventHook.OnResponse(resp)

	// TODO: As per config (set by user), but do it without conditional checks every time
	if err := decompressResponse(resp); err != nil {
		resp.Body.Close()
		return nil, fmt.Errorf("decompression failed: %w", err)
	}

	utf8reader, err := newUTF8WithFallbackReader(resp, "")
	if err != nil {
		resp.Body.Close()
		return nil, fmt.Errorf("failed to create UTF-8 reader: %w", err)
	}
	resp.Body = utf8reader

	return resp, nil
}
